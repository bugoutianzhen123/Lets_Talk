package repository

import "github.com/bugoutianzhen123/LetsTalk/repository/dao"

type Repository interface {
	User
	Group
	Friend
	FriendChat
	GroupChat
}

type repo struct {
	dao dao.Dao
}

func NewRepository(dao dao.Dao) Repository {
	return &repo{dao}
}
