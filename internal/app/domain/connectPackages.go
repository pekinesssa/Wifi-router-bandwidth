package domain

import "gorm.io/gorm"

type LoadType int
const (
	LoadTypeLow    LoadType = iota + 1 
	LoadTypeMedium                    
	LoadTypeHigh                       
)

type ConnectWifiPackages struct {
	gorm.Model
	Title string `gorm:"type:varchar(255);not null"`
	ShortDescription string `gorm:"type:text"`
	Description string `gorm:"type:text"`
	Status string `gorm:"type:varchar(255);not null;default:'gotcha'"`
	ImageUrl *string `gorm:"column:image_url;type:varchar(255);default:nil"`
	LoadType LoadType `gorm:"not null;default:1"`
	Price float64 `gorm:"type:numeric(10, 2);not null"`
	DeviceCounting int `gorm:"default:1"`
	Bandwidthconnections []Bandwidthconnections `gorm:"foreignKey:ConnectWifiPackagesID"`
}