package ioc

import (
	"database/sql"
	"fmt"

	"github.com/bugoutianzhen123/LetsTalk/pkg/logger"
	"github.com/bugoutianzhen123/LetsTalk/repository/dao"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func InitDB(l logger.Logger) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("mysql", &cfg); err != nil {
		panic(err)
	}
	// 自动建库：确保 DSN 中的数据库存在，部署时无需手工执行 CREATE DATABASE
	if err := ensureDatabase(cfg.DSN, l); err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			SlowThreshold: 0,
			LogLevel:      glogger.Info, // 以Debug模式打印所有Info级别能产生的gorm日志
		}),
	})
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

// ensureDatabase 连接 MySQL 实例（不带库名）并创建目标库（若不存在）。
// 建库失败不阻断启动：可能是库已存在但当前账号无 CREATE 权限，此时后续连接仍可成功。
func ensureDatabase(dsn string, l logger.Logger) error {
	cfg, err := mysqldriver.ParseDSN(dsn)
	if err != nil {
		return err
	}
	if cfg.DBName == "" {
		return nil
	}

	instanceCfg := *cfg
	instanceCfg.DBName = ""
	conn, err := sql.Open("mysql", instanceCfg.FormatDSN())
	if err != nil {
		return err
	}
	defer conn.Close()

	stmt := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4", cfg.DBName)
	if _, err := conn.Exec(stmt); err != nil {
		l.Warn("自动建库失败，请确认数据库已存在或有创建权限",
			logger.String("database", cfg.DBName), logger.Error(err))
		return nil
	}
	l.Info("自动建库检查完成", logger.String("database", cfg.DBName))
	return nil
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(s string, i ...interface{}) {
	g(s, logger.Field{Key: "args", Val: i})
}
