package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)


type createUserRequest struct {
	Name		string `json:"name" binding:"required"`
	Username	string `json:"username"`
	PhoneNumber string `json:"phone_number"`
	Email		string `json:"email"`
	Password	string `json:"password" binding:"required,min=8"`
	Role		string `json:"role" binding:"required,oneof=user admin superadmin system guest"`
}

type getUserRequest struct {
	UserExternalID string `uri:"userExternalID" binding:"required"`
}

type listUsersRequest struct {
	PageID		int32 `form:"page_id" binding:"required,min=1"`
	PageSize	int32 `form:"page_size" binding:"required,min=5,max=10"`
}

type updateUserRequest struct {
	UserExternalID string `uri:"userExternalID" binding:"required"`
}
type updateUserBodyRequest struct {
	Name		string `json:"name"`
	Username	string `json:"username"`
	PhoneNumber	string `json:"phone_number"`
	Email		string `json:"email"`
	Role		string `json:"role" binding:"omitempty,oneof=user admin superadmin system guest"`
}

type updatePasswordBodyRequest struct {
	Password string `json:"password" binding:"required"`
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

	arg := db.CreateUserParams{
		Name: req.Name,
		Username: pgtype.Text{String: req.Username, Valid: req.Username != ""},
		PhoneNumber: 	pgtype.Text{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
		Email:			pgtype.Text{String: req.Email, Valid: req.Email != ""},
		HashedPassword: pgtype.Text{String: hashedPassword, Valid: hashedPassword != ""},
		Role: req.Role,

	}
	/*
	arg := db.CreateUserParams{
		Name:			req.Name,
		Username:		pgtype.Text{String: req.Username, Valid: req.Username != ""},
		PhoneNumber: 	pgtype.Text{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
		Email:			pgtype.Text{String: req.Email, Valid: req.Email != ""},
		Password:		pgtype.Text{String: req.Password, Valid: req.Password != ""},
		Role:			req.Role,
	}
*/
	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (server *Server) getUser(ctx *gin.Context) {
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

	user, err := server.store.GetUser(ctx, userUUID)
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

func (server *Server) listUsers(ctx *gin.Context) {
	var req listUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListUsersParams{
		Limit: req.PageSize,
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
		Name:      bodyReq.Name,
		Username:    pgtype.Text{String: bodyReq.Username, Valid: bodyReq.Username != ""},
		PhoneNumber:  pgtype.Text{String: bodyReq.PhoneNumber, Valid: bodyReq.PhoneNumber != ""},
		Email:     pgtype.Text{String: bodyReq.Email, Valid: bodyReq.Email != ""},
		Role:      bodyReq.Role,
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
		HashedPassword:    pgtype.Text{String: hashedPassword, Valid: true},
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