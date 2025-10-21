package repository

import (
	"Lab1/internal/app/domain"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type BandwidthAdded struct{
	Address string 
	Status string
}

type CreatorBandwidthAdded struct{
	CreatedAt time.Time
	Status string
	CreatorID uint
}

type EstimateFilterOptions struct {
	Status   string
	CreatedAt *time.Time
}

type ModerateEstimateInput struct {
	Status string 
}

const (
	SimultaneousUseCoeff = 0.7 
)

var loadTypeWeights = map[domain.LoadType]float64{
	domain.LoadTypeLow:    5.0,  
	domain.LoadTypeMedium: 15.0, 
	domain.LoadTypeHigh:   25.0, 
}

type CalculationResult struct {
	TotalCost            float64
	RecommendedBandwidth float64
}

func (r* Repository) GetDraftEstimate(userID uint) (domain.BandwidthEstimate, error) {
	var estimate domain.BandwidthEstimate
	err := r.db.Preload("Bandwidthconnections").Where("creator_id = ?", userID).First(&estimate).Error
	if err != nil {
		return domain.BandwidthEstimate{}, err
	}
	return estimate, err
}

func (r *Repository) PutEstimate(id uint, inp BandwidthAdded) error{
	updatedData := make(map[string]interface{})
	if inp.Address != "" {
		updatedData["Address"] = inp.Address
	}
	if inp.Status != "" {
		updatedData["Status"] = inp.Status
	}
	
	if len(updatedData) == 0{
		return nil
	}

	result := r.db.Model(&domain.BandwidthEstimate{}).Where("id = ?", id).Updates(updatedData)
	if result.Error != nil {
		fmt.Println("Ошибка при обновлнеии")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("Заявка с ID %d не найден", id) 
	}

	return nil
}

func (r* Repository) GetFieldEstimate(userID uint) (*domain.BandwidthEstimate, error) {
	var estimate domain.BandwidthEstimate
	err := r.db.Preload("Bandwidthconnections").Where("creator_id = ?", userID).First(&estimate).Error
	if err != nil {
		return nil, err
	}

	return &estimate ,err
}

func (r *Repository) GetFilteredEstimates(filters EstimateFilterOptions) ([]domain.BandwidthEstimate, error) {
	var estimates []domain.BandwidthEstimate
	query := r.db.Model(&domain.BandwidthEstimate{})

	query = query.Joins("JOIN users ON users.id = bandwidth_estimates.creator_id")
	query = query.Where("users.is_moderator = ?", false)
	query = query.Where("bandwidth_estimates.status <> ?", "draft")
	query = query.Where("bandwidth_estimates.status <> ?", "deleted")

	if filters.Status != "" {
		query = query.Where("bandwidth_estimates.status = ?", filters.Status)
	}
	if filters.CreatedAt != nil {
		query = query.Where("bandwidth_estimates.created_at >= ?", *filters.CreatedAt)
	}

	err := query.Order("bandwidth_estimates.created_at DESC").Find(&estimates).Error
	if err != nil {
		return nil, err
	}

	return estimates, nil
}

func (r *Repository) PutCreatorEstimate(id uint, inp CreatorBandwidthAdded) error{
	updatedData := make(map[string]interface{})
	nullTime := new(time.Time)
	if inp.Status == "" {
		return fmt.Errorf("поле 'Status' является обязательным и не может быть пустым")
	}
	if inp.CreatedAt == *nullTime {
		return fmt.Errorf("поле 'CreatedAt' является обязательным и не может быть пустым")
	}
	if inp.CreatorID == 0 {
		return fmt.Errorf("поле 'Creator' является обязательным и не может быть пустым")
	}
	if inp.CreatedAt != *nullTime {
		updatedData["CreatedAt"] = inp.CreatedAt
	}
	if inp.Status != "" {
		updatedData["Status"] = inp.Status
	}
	if inp.CreatorID != 0 {
		updatedData["CreatorID"] = 1
	}
	
	if len(updatedData) == 0{
		return nil
	}

	result := r.db.Model(&domain.BandwidthEstimate{}).Where("id = ?", id).Updates(updatedData)
	if result.Error != nil {
		fmt.Println("Ошибка при обновлнеии")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("Заявка с ID %d не найден", id) 
	}

	return nil
}

func (r *Repository) ModerateEstimate(estimateID, moderatorID uint, newStatus string) (error) {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var estimate domain.BandwidthEstimate

		err := tx.Preload("Bandwidthconnections.Connection").First(&estimate, estimateID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("заявка с ID %d не найдена", estimateID)
			}
			return err
		}

		if estimate.Status == "draft" {
			return errors.New("нельзя модерировать заявку в статусе 'черновик'")
		}
		if estimate.Status == "completed" || estimate.Status == "rejected" {
			return errors.New("эта заявка уже была обработана")
		}

		updates := map[string]interface{}{
			"status":       newStatus,
			"moderator_id": moderatorID,
		}

		if newStatus == "completed" {
			totalCost, err := calculateEstimateDetails(estimate.Bandwidthconnections)
			if err != nil {
				return fmt.Errorf("ошибка при расчете стоимости: %w", err)
			}
			// updates["total_cost"] = totalCost.TotalCost
			updates["total_bandwidth"] = totalCost.RecommendedBandwidth
		}

		result := tx.Model(&estimate).Updates(updates)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})
}

func calculateEstimateDetails(connections []domain.Bandwidthconnections) (*CalculationResult, error) {
	if len(connections) == 0 {
		return nil, errors.New("нельзя завершить заявку без услуг")
	}

	var totalCost float64 = 0
	var totalBandwidthRequirement float64 = 0 

	for _, conn := range connections {
		if conn.Connection.ID == 0 {
			return nil, fmt.Errorf("детали для услуги %d не были загружены", conn.ConnectWifiPackagesID)
		}

		totalCost += conn.Connection.Price * float64(conn.DeviceCount)

		weight, ok := loadTypeWeights[conn.Connection.LoadType]
		if !ok {
			continue
		}
		
		totalBandwidthRequirement += float64(conn.DeviceCount) * weight
	}

	recommendedBandwidth := totalBandwidthRequirement * SimultaneousUseCoeff

	result := &CalculationResult{
		TotalCost:            totalCost,
		RecommendedBandwidth: recommendedBandwidth,
	}

	return result, nil
}

func (r *Repository) DeleteEstimate(estimateID uint) error {
	result := r.db.Delete(&domain.BandwidthEstimate{}, estimateID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound 
	}

	return nil
}