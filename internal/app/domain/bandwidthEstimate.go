package domain

import (
	"database/sql"
	"gorm.io/gorm"
)

type BandwidthEstimate struct {
	gorm.Model
	Status string `gorm:"type:varchar(255);not null;default:'draft'"`
	CreatorID uint `gorm:"not null"`
	Address string `gorm:"type:varchar(255)"`
	TotalBandwidth float64 `gorm:"column:total_bandwidth"`
	ModeratorID sql.NullInt64 
	
	Moderator User `gorm:"foreignKey:ModeratorID"`
	Creator User `gorm:"foreignKey:CreatorID"`
	Bandwidthconnections []Bandwidthconnections `gorm:"foreignKey:BandwidthEstimateID"`
}