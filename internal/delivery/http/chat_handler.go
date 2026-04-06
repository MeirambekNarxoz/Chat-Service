package http

import (
	"chat-service/internal/services"
	"chat-service/internal/storage"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService *services.ChatService
	minioClient *storage.MinioClient
}

func NewChatHandler(chatService *services.ChatService, minioClient *storage.MinioClient) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		minioClient: minioClient,
	}
}

func (h *ChatHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file not found in request"})
		return
	}

	fileUrl, err := h.minioClient.UploadFile(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": fileUrl})
}

func (h *ChatHandler) GetHistory(c *gin.Context) {
	chatID, err := strconv.ParseUint(c.Param("chat_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	messages, err := h.chatService.GetHistory(uint(chatID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get history"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *ChatHandler) CreateChat(c *gin.Context) {
	var req struct {
		RecipientID uint `json:"recipient_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID := c.MustGet("user_id").(uint)

	chat, err := h.chatService.CreatePersonalChat(userID, req.RecipientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create chat"})
		return
	}

	c.JSON(http.StatusCreated, chat)
}

func (h *ChatHandler) GetUserChats(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	chats, err := h.chatService.GetUserChats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get chats"})
		return
	}
	c.JSON(http.StatusOK, chats)
}
func (h *ChatHandler) CreateGroupChat(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		UserIDs []uint `json:"user_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	
	// Ensure the creator is also in the participants list
	allUserIDs := append(req.UserIDs, userID)
	
	chat, err := h.chatService.CreateGroupChat(req.Name, allUserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create group chat"})
		return
	}

	c.JSON(http.StatusCreated, chat)
}
