package repository

import (
	"Lab1/internal/models"
	"fmt"
	"os"
	"strings"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository() (*Repository, error) {
	err := godotenv.Load(".env")
	if err != nil { panic(err.Error()) }
	
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.User{}, &models.ConnectWifiPackages{}, &models.Bandwidthconnections{}, &models.BandwidthEstimate{})
	if err != nil {
		return nil, err
	}

  	return &Repository{db: db}, nil
}

// Получение всех активных услуг
func (r *Repository) GetPackages() ([]models.ConnectWifiPackages, error) {
	var Packages []models.ConnectWifiPackages
	result := r.db.Where("is_deleted = ?", false).Find(&Packages)
	return Packages, result.Error 
}

// Реализация поиска
func (r * Repository) GetPackagesByTitle(title string) ([]models.ConnectWifiPackages, error){
	var Packages []models.ConnectWifiPackages
	searchQuery := "%" + strings.ToLower(title) + "%"
	result := r.db.Where("is_deleted = ? AND lower(title) LIKE ?", false, searchQuery).Find(&Packages)
	return Packages, result.Error
}

// Получить услугу по айди 
func (r* Repository) GetPackage(id uint) (models.ConnectWifiPackages, error) {
	var Package models.ConnectWifiPackages
	result := r.db.First(&Package, id)
	if result.Error != nil {
		return models.ConnectWifiPackages{}, fmt.Errorf("Ничего не найдено, нет ") 
	}
	return Package, result.Error
}

// Получить черновик заявки
func (r* Repository) GetDraftEstimate(userID uint) (models.BandwidthEstimate, error) {
	var estimate models.BandwidthEstimate
	err := r.db.Preload("Bandwidthconnections.Connection").Where(" creator_id = ? AND status = 'draft'", userID).First(&estimate).Error
	return estimate, err
}

// Добавить услугу в заявки 
func (r * Repository) AddPackageToEstimate(userID, serviceID uint) error {
	// 1. Находим или создаем черновик для пользователя
	var estimate models.BandwidthEstimate
	if err := r.db.Where(models.BandwidthEstimate{CreatorID: userID, Status: "draft"}).FirstOrCreate(&estimate).Error; err != nil {
		return fmt.Errorf("не удалось найти или создать корзину: %w", err)
	}

	// 2. Проверяем, существует ли уже такая услуга в заявке
	var existingConnection models.Bandwidthconnections
	if err := r.db.Where("bandwidth_estimate_id = ? AND connect_wifi_packages_id = ?", estimate.ID, serviceID).First(&existingConnection).Error; err == nil {
		// Услуга уже в корзине, можно ничего не делать или увеличить количество
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("ошибка проверки наличия услуги в корзине: %w", err)
	}

	// 3. Создаем новую связь в таблице многие-ко-многим
	newConnection := models.Bandwidthconnections{
		BandwidthEstimateID:   estimate.ID,
		ConnectWifiPackagesID: serviceID,
		DeviceCount:           1,
	}

	if err := r.db.Create(&newConnection).Error; err != nil {
		return fmt.Errorf("не удалось добавить услугу в корзину: %w", err)
	}

	return nil
}

// меняем статус заявки (UPDATE)
func (r* Repository) DeleteEstimate (estimateID, userID uint) error  {
	result := r.db.Exec("UPDATE bandwidth_estimates SET status = 'deleted' WHERE id = ? AND creator_id = ?", estimateID, userID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("заявка с ID %d не найдена или у вас нет прав на ее удаление", estimateID)
	}

	return nil
}