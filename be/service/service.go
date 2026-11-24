package service

import "github.com/bugoutianzhen123/LetsTalk/repository"

type Service interface {
	UserService
	GroupService
	FriendService
	FriendChatService
	GroupChatService
}

type ser struct {
	r  repository.Repository
	cm *ClientManager
	fm *FriendManager
}

func NewService(r repository.Repository) Service {
	cm := NewClientManager(r)
	fm := NewFriendManager(r)
	return &ser{r, cm, fm}
}
