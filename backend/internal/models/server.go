package models

type Server struct {
	Model
	OwnerID         uint   `gorm:"index;not null" json:"ownerId"`
	Name            string `json:"name"`
	Address         string `json:"address"`
	OperatingSystem string `json:"operatingSystem"`
	Description     string `json:"description"`
	Status          string `gorm:"index" json:"status"`
}
