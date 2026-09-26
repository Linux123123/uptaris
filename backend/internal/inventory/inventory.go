package inventory

import (
	"context"
	"errors"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

var ErrValidation = errors.New("fields are missing, too long, or invalid")

type Page struct {
	Number int
	Size   int
}

type Filters struct {
	Status   string
	Type     string
	Severity string
}

func serverScope(db *gorm.DB, actor auth.Identity) *gorm.DB {
	if actor.Role != "admin" {
		return db.Where("owner_id = ?", actor.ID)
	}
	return db
}

func listRows[T any](query *gorm.DB, page Page, order string) ([]T, int, error) {
	var total int64
	if err := query.Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := []T{}
	err := query.Order(order).Offset((page.Number - 1) * page.Size).Limit(page.Size).Find(&rows).Error
	return rows, int(total), err
}

// update reloads and locks the row before applying and validating a patch.
// Callers supply a model previously loaded through the scoped service methods.
func update[T any](ctx context.Context, service *Service, resource *T, apply func() bool) error {
	return database.Transaction(service.db.WithContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(resource).Error; err != nil {
			return err
		}
		if !apply() {
			return ErrValidation
		}
		return tx.Model(resource).Select("*").Updates(resource).Error
	})
}
