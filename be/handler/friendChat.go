package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FriendChatHandler interface {
	HandleFriendWebSocket(c *gin.Context)
	GetFriendChatHistory(c *gin.Context)
}

// HandleFriendWebSocket 建立私聊 WebSocket 连接（按用户维度复用一条长连接）
func (h *hand) HandleFriendWebSocket(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	username := fmt.Sprintf("用户%d", userID)
	if u, err := h.ser.GetUserById(userID); err == nil {
		username = u.Username
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("私聊WebSocket升级失败: %v", err)
		return
	}

	h.ser.FriendManager().AddClient(userID, username, conn)
}

// GetFriendChatHistory 获取与某个好友的聊天记录
func (h *hand) GetFriendChatHistory(c *gin.Context) {
	userId, ok := parseUserID(c)
	if !ok {
		return
	}

	friendId, err := strconv.ParseUint(c.Query("friendid"), 10, 64)
	if err != nil || friendId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 friendid"})
		return
	}

	messages, err := h.ser.GetFriendChatHistory(userId, friendId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取聊天记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}
