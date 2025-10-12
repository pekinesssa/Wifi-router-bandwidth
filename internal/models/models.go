package models

import (
	"database/sql"
	"time"
)

// Услуги
type ConnectWifiPackages struct {
	ID uint `gorm:"primaryKey"`
	Title string `gorm:"type:varchar(255);not null"`
	ShortDescription string `gorm:"type:text"`
	Description string `gorm:"type:text"`
	IsDeleted bool   `gorm:"column:is_deleted;not null;default:false"`
	Status string `gorm:"type:varchar(255);not null;default:'gotcha'"`
	ImageUrl string `gorm:"column:image_url;type:varchar(255)"`
	Price float64 `gorm:"type:numeric(10, 2);not null"`
	DeviceCounting int `gorm:"default:1"`
	Bandwidthconnections []Bandwidthconnections `gorm:"foreignKey:ConnectWifiPackagesID"`
}

type User struct{
	ID uint `gorm:"primaryKey"`
	Login string `gorm:"type:varchar(100);not null;unique"`
	HashedPassword string `gorm:"column:hashed_password;type:varchar(255);not null"`
	IsModerator bool`gorm:"column:is_moderator;not null;default:false"`
	IsDeleted bool `gorm:"column:is_deleted;not null;default:false"`
}

// Заявкми 
type BandwidthEstimate struct {
	ID uint `gorm:"primaryKey"`
	Status string `gorm:"type:varchar(255);not null;default:'draft'"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	CreatorID uint `gorm:"not null"`
	Address string `gorm:"type:varchar(255)"`
	TotalBandwidth float64 `gorm:"column:total_bandwidth"`
	ModeratorID sql.NullInt64 
	
	Moderator User `gorm:"foreignKey:ModeratorID"`
	Creator User `gorm:"foreignKey:CreatorID"`
	Bandwidthconnections []Bandwidthconnections `gorm:"foreignKey:BandwidthEstimateID"`
}

//м-м
type Bandwidthconnections struct {
	BandwidthEstimateID uint `gorm:"primaryKey"`
    ConnectWifiPackagesID uint `gorm:"primaryKey"`

	DeviceCount int `gorm:"not null;default:1"`

	Estimate BandwidthEstimate `gorm:"foreignKey:BandwidthEstimateID"`
    Connection ConnectWifiPackages `gorm:"foreignKey:ConnectWifiPackagesID"`
}