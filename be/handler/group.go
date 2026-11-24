package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GroupHandler interface {
	CreateGroup(c *gin.Context)
	GetGroupList(c *gin.Context)
	JoinGroup(c *gin.Context)
}

func (h *hand) CreateGroup(c *gin.Context) {
	var req struct {
		Name      string `json:"name"`
		CreatorId uint64 `json:"creator_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.ser.CreateGroup(req.Name, req.CreatorId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": group})
}

func (h *hand) GetGroupList(c *gin.Context) {
	groups, err := h.ser.GetGroupList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取群聊列表失败"})
		return
	}

	// 带上 userid 时标记该用户是否已在群内，前端据此展示「加入」按钮
	joined := make(map[uint64]bool)
	if userStr := c.Query("userid"); userStr != "" {
		if userId, err := strconv.ParseUint(userStr, 10, 64); err == nil {
			if ids, err := h.ser.GetUserGroupIds(userId); err == nil {
				for _, id := range ids {
					joined[id] = true
				}
			}
		}
	}

	result := make([]gin.H, 0, len(groups))
	for _, g := range groups {
		item := gin.H{
			"id":         g.Id,
			"name":       g.Name,
			"creator_id": g.CreatorId,
			"created":    g.Created,
			"joined":     joined[g.Id],
		}
		// 已加入的群聊附带最后一条消息，供会话列表展示预览与时间
		if joined[g.Id] {
			if last, err := h.ser.GetLastGroupMessage(g.Id); err == nil {
				item["last_message"] = last.Text
				item["last_time"] = last.Created
				item["last_sender_id"] = last.SenderId
			}
		}
		result = append(result, item)
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// JoinGroup 主动加入一个已存在的群聊
func (h *hand) JoinGroup(c *gin.Context) {
	var req struct {
		GroupId uint64 `json:"group_id"`
		UserId  uint64 `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ser.JoinGroup(req.GroupId, req.UserId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已加入群聊"})
}

// getUserID 从查询参数中解析 userid
func getUserID(c *gin.Context) (uint64, bool) {
	idStr := c.Query("userid")
	if idStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "缺少 userid 参数"})
		return 0, false
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "无效的 userid"})
		return 0, false
	}
	return id, true
}
