package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zahra-pzk/Chatbot_Project3/api/dto"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)

const authorizationPayloadKey = "authorization_payload"

func (h *AuthHandler) UploadProfilePicture(c *gin.Context) {
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	payload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	extension := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s_%d%s", payload.UserExternalID, time.Now().Unix(), extension)
	filePath := filepath.Join(uploadDir, newFileName)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	fileURL := "/uploads/" + newFileName
	
	arg := db.AddPhotoToUserProfileParams{
		UserExternalID: payload.UserExternalID,
		ArrayAppend:    fileURL,
	}

	_, err = h.store.Querier.AddPhotoToUserProfile(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, dto.UploadAvatarResponse{Url: fileURL})
}

func (h *AuthHandler) GetUserProfile(c *gin.Context) {
	payload := c.MustGet(authorizationPayloadKey).(*token.Payload)
	user, err := h.store.Querier.GetUserByExternalID(c, payload.UserExternalID)
	if err != nil {
		c.JSON(http.StatusNotFound, util.ErrorResponse(err))
		return
	}

	rsp := mapUserToDTO(user)
	c.JSON(http.StatusOK, rsp)
}

func (h *AuthHandler) UpdateUserProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	payload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	user, err := h.store.Querier.GetUserByExternalID(c, payload.UserExternalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	if user.Status != db.AccountStatusDisapproved && user.Status != db.AccountStatusAwaitingVerification && user.Status != db.AccountStatusIncomplete {
		c.JSON(http.StatusForbidden, util.ErrorResponse(errors.New("profile editing is locked for verified users")))
		return
	}

	var birthDate pgtype.Date
	if req.BirthDate != "" {
		parsedTime, err := time.Parse("2006-01-02", req.BirthDate)
		if err == nil {
			birthDate = pgtype.Date{Time: parsedTime, Valid: true}
		}
	} else {
		birthDate = user.BirthDate
	}

	arg := db.UpdateUserParams{
		UserExternalID: payload.UserExternalID,
		Column2:        req.FirstName,
		Column3:        req.LastName,
		Column4:        req.Username,
		Column5:        req.Phone,
		Column6:        req.Email,
		BirthDate:      birthDate,
		Role:           string(user.Role),
		Status:         user.Status,
		Photos:         user.Photos,
	}

	updatedUser, err := h.store.Querier.UpdateUser(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, mapUserToDTO(updatedUser))
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	payload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	user, err := h.store.Querier.GetUserByExternalID(c, payload.UserExternalID)
	if err != nil {
		c.JSON(http.StatusNotFound, util.ErrorResponse(err))
		return
	}

	if err := util.CheckPassword(req.OldPassword, user.HashedPassword.String); err != nil {
		c.JSON(http.StatusUnauthorized, util.ErrorResponse(errors.New("invalid old password")))
		return
	}

	hashedPassword, err := util.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	arg := db.UpdateUserPasswordParams{
		UserExternalID: payload.UserExternalID,
		HashedPassword: pgtype.Text{String: hashedPassword, Valid: true},
	}

	err = h.store.Querier.UpdateUserPassword(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

func (h *AuthHandler) GetUserDocuments(c *gin.Context) {
	payload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	docs, err := h.store.Querier.ListDocumentsByUser(c, pgtype.UUID{Bytes: payload.UserExternalID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	var rsp []dto.UserDocumentResponse
	for _, d := range docs {
		rsp = append(rsp, dto.UserDocumentResponse{
			ID:         d.SourceExternalID.String(),
			Filename:   d.Filename.String,
			MimeType:   d.MimeType.String,
			Size:       d.SizeBytes.Int64,
			UploadedAt: d.UploadedAt.Time,
			Status:     d.Status.String,
		})
	}

	if rsp == nil {
		rsp = []dto.UserDocumentResponse{}
	}

	c.JSON(http.StatusOK, rsp)
}