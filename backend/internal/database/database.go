package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dsn string) (*gorm.DB, error) {
	databaseLogger := logger.New(log.New(os.Stderr, "", log.LstdFlags), logger.Config{
		LogLevel:                  logger.Silent,
		ParameterizedQueries:      true,
		IgnoreRecordNotFoundError: true,
	})
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: databaseLogger})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(15 * time.Minute)

	return db, nil
}
