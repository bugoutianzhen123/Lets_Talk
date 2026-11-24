package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FriendHandler interface {
	ApplyFriend(c *gin.Context)
	GetFriendApplications(c *gin.Context)
	AcceptFriendApplication(c *gin.Context)
	GetFriends(c *gin.Context)
}

// ApplyFriend 发起好友申请
func (h *hand) ApplyFriend(c *gin.Context) {
	var req struct {
		SenderId   uint64 `json:"sender_id"`
		ReceiverId uint64 `json:"receiver_id"`
		Message    string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ser.ApplyFriend(req.SenderId, req.ReceiverId, req.Message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "好友申请已发送"})
}

// GetFriendApplications 获取收到的好友申请（待处理）
func (h *hand) GetFriendApplications(c *gin.Context) {
	userId, ok := getUserID(c)
	if !ok {
		return
	}

	apps, err := h.ser.GetPendingApplications(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取好友申请失败"})
		return
	}

	result := make([]gin.H, 0, len(apps))
	for _, a := range apps {
		result = append(result, gin.H{
			"id":        a.Id,
			"message":   a.Message,
			"sender_id": a.SenderId,
			"username":  a.Sender.Username,
			"created":   a.Created,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// AcceptFriendApplication 同意好友申请
func (h *hand) AcceptFriendApplication(c *gin.Context) {
	var req struct {
		ApplicationId uint64 `json:"application_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ser.AcceptFriendApplication(req.ApplicationId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已添加为好友"})
}

// GetFriends 获取好友列表
func (h *hand) GetFriends(c *gin.Context) {
	userId, ok := getUserID(c)
	if !ok {
		return
	}

	friends, err := h.ser.GetFriends(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取好友列表失败"})
		return
	}

	result := make([]gin.H, 0, len(friends))
	for _, f := range friends {
		item := gin.H{
			"id":       f.FriendId,
			"username": f.Friend.Username,
			"online":   h.ser.IsUserOnline(f.FriendId),
		}
		// 附带与该好友的最后一条消息，供会话列表展示预览与时间
		if last, err := h.ser.GetLastFriendMessage(userId, f.FriendId); err == nil {
			item["last_message"] = last.Text
			item["last_time"] = last.Created
			item["last_sender_id"] = last.SenderId
		}
		result = append(result, item)
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
