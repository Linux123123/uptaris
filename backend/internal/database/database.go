package database

import (
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dsn string) (*gorm.DB, error) {
	connection, cockroach := strings.CutPrefix(dsn, "cockroachdb://")
	if cockroach {
		// CockroachDB speaks the PostgreSQL wire protocol, but pgx only parses PostgreSQL URL schemes.
		dsn = "postgresql://" + connection
	}

	databaseLogger := logger.New(log.New(os.Stderr, "", log.LstdFlags), logger.Config{
		LogLevel:                  logger.Silent,
		ParameterizedQueries:      true,
		IgnoreRecordNotFoundError: true,
	})
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: databaseLogger})
	if err != nil {
		return nil, err
	}

	if cockroach {
		db = db.Set("uptaris:cockroachdb", true)
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
