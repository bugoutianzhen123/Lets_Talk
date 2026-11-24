package router

import (
	"github.com/bugoutianzhen123/LetsTalk/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func InitEngine(h handler.Handler) *gin.Engine {
	g := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	g.Use(cors.New(config))

	pprof.Register(g)

	user := g.Group("/user")
	{
		user.POST("/register", h.CreateUser)
		user.POST("/login", h.Login)
		user.GET("/search", h.SearchUsers)
	}

	friend := g.Group("/friend")
	{
		friend.POST("/apply", h.ApplyFriend)
		friend.GET("/applications", h.GetFriendApplications)
		friend.POST("/accept", h.AcceptFriendApplication)
		friend.GET("/list", h.GetFriends)
		friend.GET("/history", h.GetFriendChatHistory)
	}

	group := g.Group("/group")
	{
		group.GET("/Chat", h.HandleWebSocket)
		group.GET("/history", h.GetGroupChatHistory)
		group.GET("/list", h.GetGroupList)
		group.POST("/create", h.CreateGroup)
		group.POST("/join", h.JoinGroup)
	}

	g.GET("/ws", h.HandleWebSocket)
	g.GET("/ws/friend", h.HandleFriendWebSocket)

	return g
}
