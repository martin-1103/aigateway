package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"aigateway-backend/internal/config"
	"aigateway-backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMySQL(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	// Determine log level based on environment
	logLevel := logger.Silent
	if os.Getenv("LOG_SQL") == "true" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	allModels := []interface{}{
		&models.Provider{},
		&models.Account{},
		&models.Proxy{},
		&models.ProxyStats{},
		&models.RequestLog{},
		&models.ModelMapping{},
		&models.User{},
		&models.APIKey{},
		&models.AccountQuotaPattern{},
	}

	for _, model := range allModels {
		if err := db.AutoMigrate(model); err != nil {
			// MySQL error 1061: Duplicate key name (index already exists)
			if strings.Contains(err.Error(), "1061") || strings.Contains(err.Error(), "Duplicate key name") {
				log.Printf("[Migration] Skipping duplicate index: %v", err)
				continue
			}
			return err
		}
	}
	return nil
}
