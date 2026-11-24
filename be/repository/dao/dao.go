package dao

import (
	"time"

	"github.com/bugoutianzhen123/LetsTalk/domain"
	"github.com/bugoutianzhen123/LetsTalk/pkg/logger"
	"gorm.io/gorm"
)

type Dao interface {
	UserDao
	GroupDao
	FriendDao
	friendChatDao
	groupChatDao
}

type dao struct {
	db *gorm.DB
	l  logger.Logger
}

func NewDao(db *gorm.DB, l logger.Logger) Dao {
	return &dao{db: db, l: l}
}

func InitTables(db *gorm.DB) error {
	models := []interface{}{
		&domain.UserPermission{},
		&domain.User{},

		&domain.Friend{},
		&domain.FriendChatHistory{},

		&domain.RequestStatus{},
		&domain.FriendApplication{},

		&domain.Group{},
		&domain.GroupIdentity{},
		&domain.GroupChatHistory{},
		&domain.GroupMember{},
	}

	// 逐个迁移并检查错误，避免建表失败被静默忽略
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}
	return seedData(db)
}

// seedData 写入基础种子数据：
// 1) 用户权限 2) 申请状态 3) 群身份 4) 演示群组（前端群组列表依赖固定的群组 ID）
func seedData(db *gorm.DB) error {
	permissions := []domain.UserPermission{
		{Id: 1, Permission: "user"},
		{Id: 2, Permission: "super_user"},
	}
	for _, p := range permissions {
		if err := db.Omit("Permission").FirstOrCreate(&p, domain.UserPermission{Id: p.Id}).Error; err != nil {
			return err
		}
	}

	states := []domain.RequestStatus{
		{Id: 1, State: "申请中"},
		{Id: 2, State: "已同意"},
		{Id: 3, State: "已拒绝"},
	}
	for _, st := range states {
		if err := db.Omit("State").FirstOrCreate(&st, domain.RequestStatus{Id: st.Id}).Error; err != nil {
			return err
		}
	}

	identities := []domain.GroupIdentity{
		{Id: 1, Identity: "群主"},
		{Id: 2, Identity: "管理员"},
		{Id: 3, Identity: "一般群员"},
	}
	for _, idt := range identities {
		if err := db.FirstOrCreate(&idt, domain.GroupIdentity{Id: idt.Id}).Error; err != nil {
			return err
		}
	}

	// 演示群组需要一个已存在的用户作为群主，这里确保存在一个系统账号
	var system domain.User
	if err := db.Omit("Permission").Where("username = ?", "system").First(&system).Error; err != nil {
		system = domain.User{
			Username:     "system",
			Password:     "system",
			PermissionId: 1,
			Created:      time.Now(),
			Updated:      time.Now(),
		}
		if err := db.Omit("Permission").Create(&system).Error; err != nil {
			return err
		}
	}

	// 演示群组，ID 与前端初始群组列表保持一致
	groups := []domain.Group{
		{Id: 1, Name: "Web程序设计交流群", CreatorId: system.Id},
		{Id: 2, Name: "React 讨论组", CreatorId: system.Id},
		{Id: 3, Name: "Go 后端研发组", CreatorId: system.Id},
	}
	for _, g := range groups {
		g.Created = time.Now()
		if err := db.Omit("Creator").FirstOrCreate(&g, domain.Group{Id: g.Id}).Error; err != nil {
			return err
		}
	}
	return nil
}
