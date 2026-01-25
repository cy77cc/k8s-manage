package storage

import (
	"fmt"
	"log"
	"time"

	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/logger"
	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func MustNewDB() *gorm.DB {
	var dialector gorm.Dialector
	var maxOpenConns, maxIdleConns int
	var connMaxLifetime time.Duration

	if config.CFG.MySQL.Enable {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
			config.CFG.MySQL.User,
			config.CFG.MySQL.Password,
			config.CFG.MySQL.Host,
			config.CFG.MySQL.Port,
			config.CFG.MySQL.Database,
			config.CFG.MySQL.Charset,
			config.CFG.MySQL.ParseTime,
			config.CFG.MySQL.Loc,
		)
		dialector = mysql.Open(dsn)
		maxOpenConns = config.CFG.MySQL.MaxOpenConns
		maxIdleConns = config.CFG.MySQL.MaxIdleConns
		connMaxLifetime = config.CFG.MySQL.ConnMaxLifetime
	} else if config.CFG.Postgres.Enable {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			config.CFG.Postgres.Host,
			config.CFG.Postgres.Port,
			config.CFG.Postgres.User,
			config.CFG.Postgres.Password,
			config.CFG.Postgres.Database,
			config.CFG.Postgres.SSLMode,
		)
		dialector = postgres.Open(dsn)
		maxOpenConns = config.CFG.Postgres.MaxOpenConns
		maxIdleConns = config.CFG.Postgres.MaxIdleConns
		connMaxLifetime = config.CFG.Postgres.ConnMaxLifetime
	} else if config.CFG.SQLite.Enable {
		dialector = sqlite.Open(config.CFG.SQLite.File)
		maxOpenConns = config.CFG.SQLite.MaxOpenConns
		maxIdleConns = config.CFG.SQLite.MaxIdleConns
		connMaxLifetime = config.CFG.SQLite.ConnMaxLifetime
	} else {
		// If no database is enabled, we might return nil or log error.
		// For now, let's assume one must be enabled if this is called.
		return nil
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql db: %v", err)
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	if config.CFG.App.Debug {
		db = db.Debug()
	}
	return db
}

func MustMigrate(db *gorm.DB) {
	logger.L().Info("开始迁移数据库")
	logger.L().Info(db.Migrator().CurrentDatabase())
	err := db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.AuthRefreshToken{},
	)
	if err != nil {
		log.Fatal("数据库迁移失败，请重试！！！！！")
	}
}
