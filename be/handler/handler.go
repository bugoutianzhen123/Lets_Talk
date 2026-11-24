package handler

import (
	"github.com/bugoutianzhen123/LetsTalk/pkg/logger"
	"github.com/bugoutianzhen123/LetsTalk/service"
)

type Handler interface {
	UserHandler
	GroupHandler
	FriendHandler
	FriendChatHandler
	GroupChatHandler
}

type hand struct {
	ser service.Service
	l   logger.Logger
}

func NewHandler(ser service.Service, l logger.Logger) *hand {
	return &hand{ser, l}
}
