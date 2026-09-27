package models

import (
	"context"

	"github.com/uptaris/uptaris/backend/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func update[T any](ctx context.Context, db *gorm.DB, value *T, change func(*T) error) error {
	return database.Transaction(db.WithContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(value).Error; err != nil {
			return err
		}

		if err := change(value); err != nil {
			return err
		}

		return tx.Model(value).Select("*").Updates(value).Error
	})
}

func listRows[T any](query *gorm.DB, page, pageSize int, order string) ([]T, int, error) {
	var total int64

	if err := query.Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []T{}
	err := query.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error

	return rows, int(total), err
}
