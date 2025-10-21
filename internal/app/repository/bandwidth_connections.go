package repository

import (
	"Lab1/internal/app/domain"
	"gorm.io/gorm"
)

func (r *Repository) UpdatePackageInEstimate(estimateID, packageID uint, newDeviceCount int) error {
	result := r.db.Model(&domain.Bandwidthconnections{}).
		Where("bandwidth_estimate_id = ? AND connect_wifi_packages_id = ?", estimateID, packageID).
		Update("device_count", newDeviceCount)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *Repository) DeletePackageFromEstimate(estimateID, packageID uint) error {
	result := r.db.
		Where("bandwidth_estimate_id = ? AND connect_wifi_packages_id = ?", estimateID, packageID).
		Delete(&domain.Bandwidthconnections{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}