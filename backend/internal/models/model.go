package models

import (
	"time"

	"gorm.io/gorm"
)

// Model contains database metadata shared by API resources.
// DeletedAt enables GORM soft deletion but is never exposed through JSON.
type Model struct {
	ID        uint           `gorm:"primaryKey" json:"id,string"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
