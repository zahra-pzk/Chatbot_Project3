package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zahra-pzk/Chatbot_Project3/api/dto"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/token"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)

type AuthHandler struct {
	store      *db.SQLStore
	tokenMaker token.Maker
	config     util.Config
}

func NewAuthHandler(store *db.SQLStore, tokenMaker token.Maker, config util.Config) *AuthHandler {
	return &AuthHandler{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}
}

func (h *AuthHandler) CreateGuest(c *gin.Context) {
	var req dto.CreateGuestUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}
	role := db.RoleTypeGuest
	arg := db.CreateGuestUserParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Role:      string(role),
	}
	result, err := h.store.Querier.CreateGuestUser(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	rsp := dto.CreateGuestUserResponse{
		UserExternalID: result.UserExternalID.String(),
		FirstName:      result.FirstName,
		LastName:       result.LastName,
		Email:          result.Email,
		Role:           result.Role,
		CreatedAt:      util.JalaliTime(result.CreatedAt.Time),
		UpdatedAt:      util.JalaliTime(result.UpdatedAt.Time),
	}
	c.JSON(http.StatusCreated, rsp)
}

func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	birthDate := req.BirthDate.ToTime()
	arg := db.CreateUserParams{
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Username:       pgtype.Text{String: req.Username, Valid: true},
		Email:          req.Email,
		PhoneNumber:    pgtype.Text{String: req.Phone, Valid: true},
		HashedPassword: pgtype.Text{String: hashedPassword, Valid: true},
		Role:           string(db.RoleTypeUser),
		Status:         db.AccountStatusAwaitingVerification,
		BirthDate:      pgtype.Date{Time: birthDate, Valid: true},
		Photos:         req.AvatarURLs,
	}
	result, err := h.store.CreateUserTx(c, db.CreateUserTxParams{CreateUserParams: arg})
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	rsp := dto.UserResponse{
		UserExternalID: result.User.UserExternalID.String(),
		FirstName:      result.User.FirstName,
		LastName:       result.User.LastName,
		Username:       result.User.Username.String,
		Email:          result.User.Email,
		Phone:          result.User.PhoneNumber.String,
		Role:           result.User.Role,
		Status:         string(result.User.Status),
		AvatarURLs:     result.User.Photos,
		BirthDate:      util.JalaliTime(result.User.BirthDate.Time),
		CreatedAt:      util.JalaliTime(result.User.CreatedAt.Time),
		UpdatedAt:      util.JalaliTime(result.User.UpdatedAt.Time),
	}
	c.JSON(http.StatusCreated, rsp)
}

func mapUserToDTO(u db.User) dto.UserResponse {
	return dto.UserResponse{
		UserExternalID: u.UserExternalID.String(),
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Username:       u.Username.String,
		Email:          u.Email,
		Phone:          u.PhoneNumber.String,
		Role:           u.Role,
		HashedPassword: u.HashedPassword.String,
		Status:         string(u.Status),
		AvatarURLs:     u.Photos,
		BirthDate:      util.JalaliTime(u.BirthDate.Time),
		IsOnline:       false,
		CreatedAt:      util.JalaliTime(u.CreatedAt.Time),
		UpdatedAt:      util.JalaliTime(u.UpdatedAt.Time),
		LastSeen:       util.JalaliTime(u.LastSeen.Time),
	}
}

func (h *AuthHandler) LoginUserByEmail(c *gin.Context) {
	var req dto.LoginUserWithEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}
	user, err := h.store.Querier.GetUserByEmail(c, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, util.ErrorResponse(errors.New("user not found")))
			return
		}
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	if !user.HashedPassword.Valid {
		c.JSON(http.StatusUnauthorized, util.ErrorResponse(errors.New("invalid credentials")))
		return
	}
	if err := util.CheckPassword(req.Password, user.HashedPassword.String); err != nil {
		c.JSON(http.StatusUnauthorized, util.ErrorResponse(err))
		return
	}
	accessToken, accessPayload, err := h.tokenMaker.CreateToken(
		user.UserExternalID,
		user.Username.String,
		user.Role,
		h.config.AccessTokenDuration,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	refreshToken, refreshPayload, err := h.tokenMaker.CreateToken(
		user.UserExternalID,
		user.Username.String,
		user.Role,
		h.config.RefreshTokenDuration,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	session, err := h.store.Querier.CreateSession(c, db.CreateSessionParams{
		UserExternalID: refreshPayload.UserExternalID,
		Username:       user.Username.String,
		UserAgent:      c.Request.UserAgent(),
		ClientIp:       c.ClientIP(),
		RefreshToken:   refreshToken,
		ExpiresAt:      refreshPayload.ExpiredAt,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	userDTO := mapUserToDTO(user)
	rsp := dto.LoginUserResponse{
		SessionExternalID:     session.SessionExternalID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  util.JalaliTime(accessPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: util.JalaliTime(refreshPayload.ExpiredAt),
		User:                  userDTO,
	}
	c.JSON(http.StatusOK, rsp)
}

func (h *AuthHandler) LoginUserByUsernsme(c *gin.Context) {
	var req dto.LoginUserWithUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}
	username := pgtype.Text{String: req.Username, Valid: true}
	user, err := h.store.Querier.GetUserByPhoneNumber(c, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, util.ErrorResponse(errors.New("user not found")))
			return
		}
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	if !user.HashedPassword.Valid {
		c.JSON(http.StatusUnauthorized, util.ErrorResponse(errors.New("invalid credentials")))
		return
	}
	if err := util.CheckPassword(req.Password, user.HashedPassword.String); err != nil {
		c.JSON(http.StatusUnauthorized, util.ErrorResponse(err))
		return
	}
	accessToken, accessPayload, err := h.tokenMaker.CreateToken(
		user.UserExternalID,
		user.Username.String,
		user.Role,
		h.config.AccessTokenDuration,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	refreshToken, refreshPayload, err := h.tokenMaker.CreateToken(
		user.UserExternalID,
		user.Username.String,
		user.Role,
		h.config.RefreshTokenDuration,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	session, err := h.store.Querier.CreateSession(c, db.CreateSessionParams{
		UserExternalID: refreshPayload.UserExternalID,
		Username:       user.Username.String,
		UserAgent:      c.Request.UserAgent(),
		ClientIp:       c.ClientIP(),
		RefreshToken:   refreshToken,
		ExpiresAt:      refreshPayload.ExpiredAt,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	userDTO := mapUserToDTO(user)
	rsp := dto.LoginUserResponse{
		SessionExternalID:     session.SessionExternalID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  util.JalaliTime(accessPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: util.JalaliTime(refreshPayload.ExpiredAt),
		User:                  userDTO,
	}
	c.JSON(http.StatusOK, rsp)
}

func (h *AuthHandler) LoginUserByPhone(c *gin.Context) {
	var req dto.LoginUserWithPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}
	phone := pgtype.Text{String: req.Phone, Valid: true}
	user, err := h.store.Querier.GetUserByPhoneNumber(c, phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, util.ErrorResponse(errors.New("user not found")))
			return
		}
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	if !user.HashedPassword.Valid {
		c.JSON(http.StatusUnauthorized, util.ErrorResponse(errors.New("invalid credentials")))
		return
	}
	if err := util.CheckPassword(req.Password, user.HashedPassword.String); err != nil {
		c.JSON(http.StatusUnauthorized, util.ErrorResponse(err))
		return
	}
	accessToken, accessPayload, err := h.tokenMaker.CreateToken(
		user.UserExternalID,
		user.Username.String,
		user.Role,
		h.config.AccessTokenDuration,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	refreshToken, refreshPayload, err := h.tokenMaker.CreateToken(
		user.UserExternalID,
		user.Username.String,
		user.Role,
		h.config.RefreshTokenDuration,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	session, err := h.store.Querier.CreateSession(c, db.CreateSessionParams{
		UserExternalID: refreshPayload.UserExternalID,
		Username:       user.Username.String,
		UserAgent:      c.Request.UserAgent(),
		ClientIp:       c.ClientIP(),
		RefreshToken:   refreshToken,
		ExpiresAt:      refreshPayload.ExpiredAt,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	userDTO := mapUserToDTO(user)
	rsp := dto.LoginUserResponse{
		SessionExternalID:     session.SessionExternalID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  util.JalaliTime(accessPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: util.JalaliTime(refreshPayload.ExpiredAt),
		User:                  userDTO,
	}
	c.JSON(http.StatusOK, rsp)
}
