package model

import (
	"database/sql"
	"os"
	"time"

	"github.com/kingfer30/topup-online/constants"
	crypto "github.com/kingfer30/topup-online/utils/cypto"
	"github.com/kingfer30/topup-online/utils/env"
	"github.com/kingfer30/topup-online/utils/logger"
	"github.com/kingfer30/topup-online/utils/random"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var LOG_DB *gorm.DB

func CreateRootAccountIfNeed() error {
	var user User
	//if user.Status != util.UserStatusEnabled {
	if err := DB.First(&user).Error; err != nil {
		logger.SysLog("no user exists, creating a root user for you: username is root, password is 123456")
		hashedPassword, err := crypto.Password2Hash("123456")
		if err != nil {
			return err
		}
		accessToken := random.GetUUID()
		rootUser := User{
			Username:    "root",
			Password:    hashedPassword,
			Role:        constants.RoleRootUser,
			Status:      constants.UserStatusEnabled,
			DisplayName: "Root User",
			AccessToken: accessToken,
			Quota:       500000000000000,
		}
		DB.Create(&rootUser)
	}
	return nil
}

func openMySQL(dsn string) (*gorm.DB, error) {
	logger.SysLog("using MySQL as database")
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt: true, // precompile SQL
	})
}

func InitDB() {
	dsn := os.Getenv("SQL_DSN")
	var err error
	DB, err = openMySQL(dsn)
	if err != nil {
		logger.FatalLog("failed to initialize database: " + err.Error())
		return
	}

	_ = setDBConns(DB)

	logger.SysLog("database migration started")
	if err = migrateDB(); err != nil {
		logger.FatalLog("failed to migrate database: " + err.Error())
		return
	}
	logger.SysLog("database migrated")
}

func migrateDB() error {
	var err error
	if err = DB.AutoMigrate(&User{}); err != nil {
		return err
	}
	return nil
}

func setDBConns(db *gorm.DB) *sql.DB {
	if constants.DebugSQLEnabled {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.FatalLog("failed to connect database: " + err.Error())
		return nil
	}

	sqlDB.SetMaxIdleConns(env.Int("SQL_MAX_IDLE_CONNS", 100))
	sqlDB.SetMaxOpenConns(env.Int("SQL_MAX_OPEN_CONNS", 1000))
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(env.Int("SQL_MAX_LIFETIME", 60)))
	return sqlDB
}

func closeDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}

func CloseDB() error {
	if LOG_DB != DB {
		err := closeDB(LOG_DB)
		if err != nil {
			return err
		}
	}
	return closeDB(DB)
}
