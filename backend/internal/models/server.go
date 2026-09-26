package models

import (
	"context"

	"github.com/uptaris/uptaris/backend/internal/database"
	"gorm.io/gorm"
)

type Server struct {
	Model
	OwnerID         uint   `gorm:"index;not null" json:"ownerId,string"`
	Name            string `json:"name" binding:"required,max=200"`
	Address         string `json:"address" binding:"required,max=2048"`
	OperatingSystem string `json:"operatingSystem" binding:"required,max=200"`
	Description     string `json:"description" binding:"max=10000"`
	Status          string `gorm:"index" json:"status" binding:"required,oneof=up down paused"`
}

func (value *Server) Create(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Create(value).Error
}

func (value *Server) Update(ctx context.Context, db *gorm.DB, change func(*Server) error) error {
	return update(ctx, db, value, change)
}

func GetServer(ctx context.Context, db *gorm.DB, ownerID uint, admin bool, id uint) (*Server, error) {
	value := new(Server)
	err := serverScope(db.WithContext(ctx), ownerID, admin).First(value, id).Error
	return value, err
}

func ListServers(ctx context.Context, db *gorm.DB, ownerID uint, admin bool, page Page, filters Filters) ([]Server, int, error) {
	query := serverScope(db.WithContext(ctx), ownerID, admin)
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	return listRows[Server](query, page, "created_at desc, id desc")
}

func serverScope(db *gorm.DB, ownerID uint, admin bool) *gorm.DB {
	if !admin {
		return db.Where("owner_id = ?", ownerID)
	}
	return db
}

func (value *Server) Delete(ctx context.Context, db *gorm.DB) error {
	return database.Transaction(db.WithContext(ctx), func(tx *gorm.DB) error {
		monitors := tx.Model(&Monitor{}).Select("id").Where("server_id = ?", value.ID)
		if err := tx.Where("monitor_id IN (?)", monitors).Delete(&Incident{}).Error; err != nil {
			return err
		}
		if err := tx.Where("server_id = ?", value.ID).Delete(&Monitor{}).Error; err != nil {
			return err
		}
		return tx.Delete(value).Error
	})
}
