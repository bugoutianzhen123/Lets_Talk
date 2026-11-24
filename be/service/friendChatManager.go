package service

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/bugoutianzhen123/LetsTalk/domain"
	"github.com/bugoutianzhen123/LetsTalk/repository"
	"github.com/gorilla/websocket"
)

type FriendUser struct {
	Conn     *websocket.Conn
	UserId   uint64
	Username string
}

// FriendMessage 私聊消息（前端根据 sender_id 与本机用户比较判断收发方向）
type FriendMessage struct {
	SenderId uint64    `json:"sender_id"`
	Username string    `json:"username"`
	TargetId uint64    `json:"target_id"`
	Content  string    `json:"content"`
	SendTime time.Time `json:"send_time"`
}

// FriendManager 按用户维度维护连接，负责私聊消息的存储与投递
type FriendManager struct {
	r     repository.FriendChat
	users map[uint64][]*FriendUser
	mutex sync.Mutex
}

func NewFriendManager(r repository.FriendChat) *FriendManager {
	return &FriendManager{
		r:     r,
		users: make(map[uint64][]*FriendUser),
	}
}

// AddClient 注册一个用户连接（同一用户可多端在线）
func (fm *FriendManager) AddClient(userId uint64, username string, conn *websocket.Conn) {
	fm.mutex.Lock()
	fm.users[userId] = append(fm.users[userId], &FriendUser{
		Conn:     conn,
		UserId:   userId,
		Username: username,
	})
	fm.mutex.Unlock()

	go fm.handleClient(userId, username, conn)
}

// RemoveClient 移除连接
func (fm *FriendManager) RemoveClient(userId uint64, conn *websocket.Conn) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	list := fm.users[userId]
	next := list[:0]
	for _, u := range list {
		if u.Conn != conn {
			next = append(next, u)
		}
	}
	if len(next) == 0 {
		delete(fm.users, userId)
	} else {
		fm.users[userId] = next
	}
	conn.Close()
}

// IsOnline 该用户是否存在活跃的私聊长连接
func (fm *FriendManager) IsOnline(userId uint64) bool {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	return len(fm.users[userId]) > 0
}

func (fm *FriendManager) handleClient(userId uint64, username string, conn *websocket.Conn) {
	defer fm.RemoveClient(userId, conn)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("私聊客户端异常断开: %v", err)
			}
			return
		}

		var payload struct {
			TargetId uint64 `json:"target_id"`
			Content  string `json:"content"`
		}
		if err := json.Unmarshal(message, &payload); err != nil {
			log.Printf("私聊消息解析失败: %v", err)
			continue
		}
		if payload.TargetId == 0 || payload.Content == "" {
			continue
		}

		msg := FriendMessage{
			SenderId: userId,
			Username: username,
			TargetId: payload.TargetId,
			Content:  payload.Content,
			SendTime: time.Now(),
		}

		// 先落库，保证对方离线时也能通过历史接口拿到消息
		if err := fm.r.SaveFriendMessage(domain.FriendChatHistory{
			SenderId: userId,
			TargetId: payload.TargetId,
			Text:     payload.Content,
			Created:  msg.SendTime,
		}); err != nil {
			log.Printf("私聊消息保存失败: %v", err)
		}

		fm.deliver(msg)
	}
}

// deliver 将消息投递给发送方（多端同步）与接收方
func (fm *FriendManager) deliver(msg FriendMessage) {
	data, _ := json.Marshal(msg)

	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	seen := make(map[uint64]bool)
	for _, uid := range []uint64{msg.SenderId, msg.TargetId} {
		if seen[uid] {
			continue
		}
		seen[uid] = true
		for _, u := range fm.users[uid] {
			if err := u.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("私聊消息发送失败: %v", err)
				u.Conn.Close()
			}
		}
	}
}
