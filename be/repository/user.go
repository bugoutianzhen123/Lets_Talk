package repository

import "github.com/bugoutianzhen123/LetsTalk/domain"

type User interface {
	CreateUser(user domain.User) error
	GetUserByUsername(username string) (domain.User, error)
	GetUserById(id uint64) (domain.User, error)
	SearchUsers(keyword string) ([]domain.User, error)
}

func (r *repo) CreateUser(user domain.User) error {
	return r.dao.CreateUser(user)
}

func (r *repo) GetUserByUsername(username string) (domain.User, error) {
	return r.dao.GetUserByUsername(username)
}

func (r *repo) GetUserById(id uint64) (domain.User, error) {
	return r.dao.GetUserById(id)
}

func (r *repo) SearchUsers(keyword string) ([]domain.User, error) {
	return r.dao.SearchUsers(keyword)
}
