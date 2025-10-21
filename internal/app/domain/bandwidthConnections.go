package domain
import()

type Bandwidthconnections struct {
	BandwidthEstimateID uint `gorm:"primaryKey"`
    ConnectWifiPackagesID uint `gorm:"primaryKey"`

	DeviceCount int `gorm:"not null;default:1"`

	Estimate BandwidthEstimate `gorm:"foreignKey:BandwidthEstimateID"`
    Connection ConnectWifiPackages `gorm:"foreignKey:ConnectWifiPackagesID"`
}