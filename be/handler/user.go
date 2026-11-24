package handler

import (
	"github.com/bugoutianzhen123/LetsTalk/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserHandler interface {
	CreateUser(c *gin.Context)
	Login(c *gin.Context)
	SearchUsers(c *gin.Context)
}

func (h *hand) CreateUser(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ser.CreateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "注册失败：该用户名可能已被占用"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (h *hand) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名和密码"})
		return
	}

	user, err := h.ser.Login(req.Username, req.Password)
	if err != nil {
		// 账号不存在 / 密码错误均返回中文提示
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"user": gin.H{
			"id":       user.Id,
			"username": user.Username,
		},
	})
}

// SearchUsers 按用户名关键词搜索用户，仅返回 id 与用户名，避免泄露密码等信息
func (h *hand) SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")

	users, err := h.ser.SearchUsers(keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}

	result := make([]gin.H, 0, len(users))
	for _, u := range users {
		result = append(result, gin.H{
			"id":       u.Id,
			"username": u.Username,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
