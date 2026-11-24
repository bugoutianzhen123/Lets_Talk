package service

import (
	"errors"
	"github.com/bugoutianzhen123/LetsTalk/domain"
	"time"
)

type UserService interface {
	CreateUser(user domain.User) error
	Login(username, password string) (domain.User, error)
	GetUserById(id uint64) (domain.User, error)
	SearchUsers(keyword string) ([]domain.User, error)
}

var (
	// ErrUserNotFound 账号不存在
	ErrUserNotFound = errors.New("该账号不存在，请先注册")
	// ErrWrongPassword 密码错误
	ErrWrongPassword = errors.New("密码错误，请重新输入")
)

func (s *ser) CreateUser(user domain.User) error {
	user.Created = time.Now()
	user.Updated = user.Created
	// 新注册用户默认为普通用户（user_permissions 表中 id=1 为 user）
	if user.PermissionId == 0 {
		user.PermissionId = 1
	}
	return s.r.CreateUser(user)
}

func (s *ser) Login(username, password string) (domain.User, error) {
	user, err := s.r.GetUserByUsername(username)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}
	if user.Password != password {
		return domain.User{}, ErrWrongPassword
	}
	return user, nil
}

func (s *ser) GetUserById(id uint64) (domain.User, error) {
	return s.r.GetUserById(id)
}

func (s *ser) SearchUsers(keyword string) ([]domain.User, error) {
	return s.r.SearchUsers(keyword)
}
