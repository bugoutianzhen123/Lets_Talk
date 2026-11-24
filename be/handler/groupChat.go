package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/bugoutianzhen123/LetsTalk/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type GroupChatHandler interface {
	HandleWebSocket(c *gin.Context)
	GetGroupChatHistory(c *gin.Context)
}

type groupChatHandler struct {
	clientManager *service.ClientManager
}

func NewGroupChatHandler(cm *service.ClientManager) *groupChatHandler {
	return &groupChatHandler{clientManager: cm}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *hand) HandleWebSocket(c *gin.Context) {
	groupID, ok1 := parseGroupID(c)
	userID, ok2 := parseUserID(c)
	if !ok1 || !ok2 {
		return
	}

	// 查询用户信息，用于在聊天中展示注册用户名
	username := fmt.Sprintf("用户%d", userID)
	if u, err := h.ser.GetUserById(userID); err == nil {
		username = u.Username
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	gm := h.ser.FindGroup(groupID)
	gm.AddClient(userID, username, conn)
}

func (h *hand) GetGroupChatHistory(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 100 {
		limit = 50
	}

	messages, err := h.ser.GetGroupChatHistory(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// 新增：转换 groupID 的辅助函数
func parseGroupID(c *gin.Context) (uint64, bool) {
	groupStr := c.Query("groupid")
	if groupStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "缺少 groupid 参数",
		})
		return 0, false
	}

	groupID, err := strconv.ParseUint(groupStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("无效的 groupid: %s", groupStr),
		})
		return 0, false
	}
	return groupID, true
}

func parseUserID(c *gin.Context) (uint64, bool) {
	userStr := c.Query("userid")
	if userStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "缺少 userid 参数",
		})
		return 0, false
	}

	userID, err := strconv.ParseUint(userStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("无效的 userid: %s", userStr),
		})
		return 0, false
	}
	return userID, true
}
