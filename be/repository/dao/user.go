package dao

import "github.com/bugoutianzhen123/LetsTalk/domain"

type UserDao interface {
	CreateUser(u domain.User) error
	GetUserByUsername(username string) (domain.User, error)
	GetUserById(id uint64) (domain.User, error)
	SearchUsers(keyword string) ([]domain.User, error)
}

func (d *dao) CreateUser(u domain.User) error {
	return d.db.Create(&u).Error
}

func (d *dao) GetUserByUsername(username string) (domain.User, error) {
	var u domain.User
	err := d.db.Where("username = ?", username).First(&u).Error
	return u, err
}

func (d *dao) GetUserById(id uint64) (domain.User, error) {
	var u domain.User
	err := d.db.Where("id = ?", id).First(&u).Error
	return u, err
}

func (d *dao) SearchUsers(keyword string) ([]domain.User, error) {
	var users []domain.User
	// 隐藏系统账号，按用户名模糊匹配
	err := d.db.Where("username LIKE ? AND username != ?", "%"+keyword+"%", "system").
		Limit(20).
		Find(&users).Error
	return users, err
}
