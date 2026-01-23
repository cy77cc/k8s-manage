package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func MustNewDB() *gorm.DB {
	var dialector gorm.Dialector
	var maxOpenConns, maxIdleConns int
	var connMaxLifetime time.Duration

	if CFG.MySQL.Enable {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
			CFG.MySQL.User,
			CFG.MySQL.Password,
			CFG.MySQL.Host,
			CFG.MySQL.Port,
			CFG.MySQL.Database,
			CFG.MySQL.Charset,
			CFG.MySQL.ParseTime,
			CFG.MySQL.Loc,
		)
		dialector = mysql.Open(dsn)
		maxOpenConns = CFG.MySQL.MaxOpenConns
		maxIdleConns = CFG.MySQL.MaxIdleConns
		connMaxLifetime = CFG.MySQL.ConnMaxLifetime
	} else if CFG.Postgres.Enable {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			CFG.Postgres.Host,
			CFG.Postgres.Port,
			CFG.Postgres.User,
			CFG.Postgres.Password,
			CFG.Postgres.Database,
			CFG.Postgres.SSLMode,
		)
		dialector = postgres.Open(dsn)
		maxOpenConns = CFG.Postgres.MaxOpenConns
		maxIdleConns = CFG.Postgres.MaxIdleConns
		connMaxLifetime = CFG.Postgres.ConnMaxLifetime
	} else if CFG.SQLite.Enable {
		dialector = sqlite.Open(CFG.SQLite.File)
		maxOpenConns = CFG.SQLite.MaxOpenConns
		maxIdleConns = CFG.SQLite.MaxIdleConns
		connMaxLifetime = CFG.SQLite.ConnMaxLifetime
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

	if CFG.App.Debug {
		db = db.Debug()
	}

	return db
}
