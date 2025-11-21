package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/zahra-pzk/Chatbot_Project3/db/sqlc"
)

type startChatRequest struct {
	UserExternalID	uuid.UUID	`json:"user_external_id" binding:"required"`
	Content			string		`json:"content" binding:"required"`
}

func (server *Server) createChat(ctx *gin.Context) {
	var req startChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.StartChatTxParams{
		UserExternalID: req.UserExternalID,
		Content: req.Content,
	}

	result, err := server.store.CreateChatTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	//chat, err := server.store.CreateChat(ctx, arg)
	//if err != nil {
	//	ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	//	return
	//}
	ctx.JSON(http.StatusOK, result)
}

type getChatRequest struct {
	ChatExternalID string `uri:"chatExternalID" binding:"required"`
}

func (server *Server) getChat(ctx *gin.Context) {
	var req getChatRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	chatUUID, err := uuid.Parse(req.ChatExternalID)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }
	chat, err := server.store.GetChat(ctx, chatUUID)

	if err != nil {
    	if err == sql.ErrNoRows {
    		ctx.JSON(http.StatusNotFound, errorResponse(err))
    		return
    	}
    	ctx.JSON(http.StatusInternalServerError, errorResponse(err))
    	return
  	}
  	ctx.JSON(http.StatusOK, chat)
}
type getChatsByUserRequest struct {
	UserExternalID	string `form:"user_external_id" binding:"required"`
	pageID			int32 `form:"page_id" binding:"required,min=1"`
	pageSize		int32 `form:"page_size" binding:"required,min=5,max=10"`

}


func (server *Server) getChatsByUser(ctx *gin.Context) {
	var req getChatsByUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userUUID, err := uuid.Parse(req.UserExternalID)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

	arg := db.GetChatsByUserParams {
		UserExternalID: userUUID,
		Limit: req.pageSize,
		Offset: (req.pageID - 1) * req.pageSize,
	}

	chats, err := server.store.GetChatsByUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, chats)
}

type listChatRequest struct {
	pageID		int32 `form:"page_id" binding:"required,min=1"`
	pageSize	int32 `form:"page_size" binding:"required,min=5,max=10"`

}

func (server *Server) listChats(ctx *gin.Context) {
	var req listChatRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListChatsParams {
		Limit: req.pageSize,
		Offset: (req.pageID - 1) * req.pageSize,
	}

	chats, err := server.store.ListChats(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, chats)
}

 type deleteChatRequest struct {
	ChatExternalID string `uri:"chatExternalID" binding:"required"`
}

func (server *Server) deleteChat(ctx *gin.Context) {
  var req deleteChatRequest
  if err := ctx.ShouldBindUri(&req); err != nil {
    ctx.JSON(http.StatusBadRequest, errorResponse(err))
    return
  }

    chatUUID, err := uuid.Parse(req.ChatExternalID)
  if err != nil {
    ctx.JSON(http.StatusBadRequest, errorResponse(err))
    return
  }

  err = server.store.DeleteChat(ctx, chatUUID)
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

type updateChatStatusRequest struct {
	ChatExternalID	string `uri:"chatExternalID" binding:"required"`
}

type updateChatStatusBodyRequest struct {
	Status string `json:"status" binding:"required,oneof=pending open closed"`
}


func (server *Server) updateChatStatus(ctx *gin.Context) {
	var uriReq updateChatStatusRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil { 
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	chatUUID, err := uuid.Parse(uriReq.ChatExternalID)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

	var bodyReq updateChatStatusBodyRequest
	if err := ctx.ShouldBindJSON(&bodyReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateChatStatusParams{
		ChatExternalID: chatUUID,
		Status:			bodyReq.Status,
	}

	chat, err := server.store.UpdateChatStatus(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, chat) 
}

type updateChatRequest struct {
	ChatExternalID	string `uri:"chatExternalID" binding:"required"`
}

type updateChatBodyRequest struct {
	UserExternalID	string		`json:"user_external_id"`
	Status			string		`json:"status"`
}

func (server *Server) updateChat(ctx *gin.Context) {
	var uriReq updateChatRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil { 
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	chatUUID, err := uuid.Parse(uriReq.ChatExternalID)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }
	var bodyReq updateChatBodyRequest
	if err := ctx.ShouldBindJSON(&bodyReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	var userUUID uuid.UUID
	if bodyReq.UserExternalID != "" {
		userUUID, err = uuid.Parse(bodyReq.UserExternalID)
    	if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
    	}
	}
	arg := db.UpdateChatParams{
		ChatExternalID: chatUUID,
		Column2: userUUID, 
		Status:			bodyReq.Status,
	}
	chat, err := server.store.UpdateChat(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, chat) 
}