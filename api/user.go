package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)

type createUserRequest struct {
	Name        string `json:"name" binding:"required"`
	Username    string `json:"username" binding:"required"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	Password    string `json:"password" binding:"required,min=8"`
	Role        string `json:"role" binding:"required,oneof=user admin superadmin system guest"`
}

type userResponse struct {
	ExternalID  uuid.UUID `json:"user_external_id"`
	Name        string    `json:"name"`
	Username    string    `json:"username"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
	BirthDate   time.Time `json:"birth_date"`
	Photos      []string  `json:"photos"`
}

type getUserRequest struct {
	UserExternalID string `uri:"userExternalID" binding:"required"`
}

type listUsersRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

type updateUserRequest struct {
	UserExternalID string `uri:"userExternalID" binding:"required"`
}
type updateUserBodyRequest struct {
	Name        string `json:"name"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	Role        string `json:"role" binding:"omitempty,oneof=user admin superadmin system guest"`
}

type updatePasswordBodyRequest struct {
	Password string `json:"password" binding:"required"`
}

func newUserResponse(user db.User) userResponse {
	var username, phone, email string
	var createdAt time.Time
	if user.Username.Valid {
		username = user.Username.String
	}
	if user.PhoneNumber.Valid {
		phone = user.PhoneNumber.String
	}
	if user.Email.Valid {
		email = user.Email.String
	}
	if user.CreatedAt.Valid {
		createdAt = user.CreatedAt.Time
	}
	return userResponse{
		ExternalID:  user.UserExternalID,
		Name:        user.Name,
		Username:    username,
		PhoneNumber: phone,
		Email:       email,
		Role:        user.Role,
		CreatedAt:   createdAt,
	}
}

func newUserProfileResponse(data db.GetUserAccountByExternalIDRow) userResponse {
	var username, phone, email string
	if data.Username.Valid {
		username = data.Username.String
	}

	if data.PhoneNumber.Valid {
		phone = data.PhoneNumber.String
	}
	if data.Email.Valid {
		email = data.Email.String
	}
	var birthDate time.Time
	if data.BirthDate.Valid {
		birthDate = data.BirthDate.Time
	}

	var createdAt time.Time
	if data.CreatedAt.Valid {
		createdAt = data.CreatedAt.Time
	}
	return userResponse{
		ExternalID:  data.AccountExternalID,
		Name:        data.Name,
		Username:    username,
		PhoneNumber: phone,
		Email:       email,
		Role:        data.Role,
		Status:      string(data.Status),
		BirthDate:   birthDate,
		Photos:      data.Photos,
		CreatedAt:   createdAt,
	}
}
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Name:           req.Name,
			Username:       pgtype.Text{String: req.Username, Valid: req.Username != ""},
			PhoneNumber:    pgtype.Text{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
			Email:          pgtype.Text{String: req.Email, Valid: req.Email != ""},
			HashedPassword: pgtype.Text{String: hashedPassword, Valid: hashedPassword != ""},
			Role:           req.Role,
		},
	}

	result, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	rsp := newUserResponse(result.User)
	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) getUser(ctx *gin.Context) {
	idStr := ctx.Param("userExternalID")
	userUUID, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	userProfile, err := server.store.GetUserAccountByExternalID(ctx, userUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	rsp := newUserProfileResponse(userProfile)
	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) listUsers(ctx *gin.Context) {
	var req listUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListUsersParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	users, err := server.store.ListUsers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, users)
}

func (server *Server) deleteUser(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	userUUID, err := uuid.Parse(req.UserExternalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = server.store.DeleteUser(ctx, userUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}

func (server *Server) updateUser(ctx *gin.Context) {
	var uriReq updateUserRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	userUUID, err := uuid.Parse(uriReq.UserExternalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var bodyReq updateUserBodyRequest
	if err := ctx.ShouldBindJSON(&bodyReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateUserParams{
		UserExternalID: userUUID,
		Name:           bodyReq.Name,
		Username:       pgtype.Text{String: bodyReq.Username, Valid: bodyReq.Username != ""},
		PhoneNumber:    pgtype.Text{String: bodyReq.PhoneNumber, Valid: bodyReq.PhoneNumber != ""},
		Email:          pgtype.Text{String: bodyReq.Email, Valid: bodyReq.Email != ""},
		Role:           bodyReq.Role,
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (server *Server) updatePassword(ctx *gin.Context) {
	var uriReq updateUserRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	userUUID, err := uuid.Parse(uriReq.UserExternalID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var bodyReq updatePasswordBodyRequest
	if err := ctx.ShouldBindJSON(&bodyReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hashedPassword, err := util.HashPassword(bodyReq.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.UpdateUserPasswordParams{
		UserExternalID: userUUID,
		HashedPassword: pgtype.Text{String: hashedPassword, Valid: true},
	}

	err = server.store.UpdateUserPassword(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginUserResponse struct {
	AccessToken string       `json:"access_token"`
	User        userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByUsername(ctx, pgtype.Text{String: req.Username, Valid: req.Username != ""})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if !user.HashedPassword.Valid {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("invalid credentials")))
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword.String)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(
		user.UserExternalID,
		user.Username.String,
		user.Role,
		server.config.AccessTokenDuration,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, rsp)
}
