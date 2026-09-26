package database

import (
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbgorm"
	"gorm.io/gorm"
)

// Transaction retries CockroachDB serialization failures while preserving GORM transactions for PostgreSQL.
func Transaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	if cockroach, _ := db.Get("uptaris:cockroachdb"); cockroach == true {
		return crdbgorm.ExecuteTx(db.Statement.Context, db, nil, fn)
	}
	return db.Transaction(fn)
}
