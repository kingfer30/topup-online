package model

import (
	"database/sql"
	"os"
	"time"

	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/utils/env"
	"github.com/kingfer30/topup-online/utils/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var LOG_DB *gorm.DB

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
	// 迁移所有表
	if err = DB.AutoMigrate(&User{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&MirrorCard{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&SystemConfig{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&Admin{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&AdminSession{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&Order{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&Menu{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&SalesTalk{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&DigisellerOrder{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&DigisellerPrice{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&GptCard{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&GptCdk{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&GptTopupTask{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&GptRtLicense{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&GptRtLicenseDevice{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&GptRtLicenseHold{}); err != nil {
		return err
	}
	if err = MigrateGptRtLicenseLegacyMachineID(); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&MicrosoftMail{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&AdConfig{}); err != nil {
		return err
	}
	// 对所有 cards_* 动态表执行新增列迁移
	if err = MigrateCardTableColumns(); err != nil {
		return err
	}
	return nil
}

// MigrateDB 公开的数据库迁移方法，供外部调用
func MigrateDB() error {
	return migrateDB()
}

func setDBConns(db *gorm.DB) *sql.DB {
	if constants.GetDebugSQLEnabled() {
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
