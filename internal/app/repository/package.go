package repository

import (
	"Lab1/internal/app/domain"
	"Lab1/internal/pkg"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
	minioClient *pkg.Client
	minioPublicURL string
	
}

func NewRepository(minioClient *pkg.Client, minioPublicURL string) (*Repository, error) {
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

	err = db.AutoMigrate(&domain.User{}, &domain.ConnectWifiPackages{}, &domain.Bandwidthconnections{}, &domain.BandwidthEstimate{})
	if err != nil {
		return nil, err
	}

  	return &Repository{
		db: db,
		minioClient: minioClient,
		minioPublicURL: minioPublicURL,
	}, nil
}

// Получение всех активных услуг
func (r *Repository) GetPackages() ([]domain.ConnectWifiPackages, error) {
	var Packages []domain.ConnectWifiPackages
	result := r.db.Where("is_deleted = ?", false).Find(&Packages)
	return Packages, result.Error 
}

// Реализация поиска
func (r * Repository) GetPackagesByTitle(title string) ([]domain.ConnectWifiPackages, error){
	var Packages []domain.ConnectWifiPackages
	searchQuery := "%" + strings.ToLower(title) + "%"
	result := r.db.Where("is_deleted = ? AND lower(title) LIKE ?", false, searchQuery).Find(&Packages)
	return Packages, result.Error
}

// Получить услугу по айди 
func (r* Repository) GetPackage(id uint) (domain.ConnectWifiPackages, error) {
	var Package domain.ConnectWifiPackages
	result := r.db.First(&Package, id)
	if result.Error != nil {
		return domain.ConnectWifiPackages{}, fmt.Errorf("Ничего не найдено, нет ") 
	}
	return Package, result.Error
}

func (r *Repository) PostPackage(imp CreatePackage) (*domain.ConnectWifiPackages, error) {
	newPackage := &domain.ConnectWifiPackages{
		Title: imp.Title, 
		// IsDeleted: false, 
		Status: imp.Status, 
		Price: imp.Price, 
		DeviceCounting: imp.DeviceCounting,
	}

	result := r.db.Create(newPackage)
	if result.Error != nil {
		fmt.Println("Ошибка при записи")
		return nil, result.Error
	}	
	return newPackage, result.Error
}

func (r *Repository) PutPackage(id uint, inp UpdatePackage) error{
	updatedData := make(map[string]interface{})
	if inp.Title != "" {
		updatedData["Title"] = inp.Title
	}
	if inp.Description != "" {
		updatedData["Description"] = inp.Description
	}
	if inp.ShortDescription != "" {
		updatedData["ShortDescription"] = inp.ShortDescription
	}
	if inp.DeviceCounting != 0 {
		updatedData["DeviceCounting"] = inp.DeviceCounting
	}
	if inp.Price != 0 {
		updatedData["Price"] = inp.Price
	}
	if inp.Status != "" {
		updatedData["Status"] = inp.Status
	}

	if len(updatedData) == 0{
		return nil
	}

	result := r.db.Model(&domain.ConnectWifiPackages{}).Where("id = ?", id).Updates(updatedData)
	if result.Error != nil {
		fmt.Println("Ошибка при обновлнеии")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("пакет с ID %d не найден", id) 
	}

	return nil
}

func (r *Repository) DeletePackage(id uint) error{
	var delete domain.ConnectWifiPackages
	if err := r.db.First(&delete, id).Error; err != nil {
		return fmt.Errorf("Не удалось пакет с такоим айди", id) 
	}

	imageUrl := delete.ImageUrl
	
	if imageUrl != nil {
		cleanStr := strings.TrimPrefix(*imageUrl, "http://localhost:9000/main/")
		err := r.minioClient.DeleteObject(context.Background(), cleanStr)
			if err != nil {
				logrus.Errorf(
					"КРИТИЧЕСКАЯ ОШИБКА: Запись из БД удалена, но не удалось удалить объект '%s' из MinIO: %v",
					*imageUrl,
					err,
				)
			} else {
				logrus.Infof("Объект '%s' успешно удален из MinIO.", imageUrl)
			}
	}
	
	result:=r.db.Delete(&delete)
	if result.Error != nil {
		logrus.Errorf("Не удалось удалить пакет", result.Error)
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("Не удалось пакет с такоим айди", result.Error) 
	}
	
	return nil
}

// Добавить услугу в заявки 
func (r * Repository) AddPackageToEstimate(userID, serviceID uint, devices int) (*domain.Bandwidthconnections, error) {
	var estimate domain.BandwidthEstimate
	if err := r.db.Where(domain.BandwidthEstimate{CreatorID: userID, Status: "draft"}).FirstOrCreate(&estimate).Error; err != nil {
		return nil, fmt.Errorf("не удалось найти или создать корзину: %w", err)
	}

	var existingConnection domain.Bandwidthconnections
	err := r.db.Where("bandwidth_estimate_id = ? AND connect_wifi_packages_id = ?", estimate.ID, serviceID).First(&existingConnection).Error
	if err == nil {
		if updateErr := r.db.Model(&existingConnection).Update("device_count", devices).Error; updateErr != nil {
			return nil, fmt.Errorf("не удалось обновить услугу в черновике: %w", updateErr)
		}

		return &existingConnection, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("ошибка проверки наличия услуги в корзине: %w", err)
	}

	newConnection := domain.Bandwidthconnections{
		BandwidthEstimateID:   estimate.ID,
		ConnectWifiPackagesID: serviceID,
		DeviceCount:           devices,
	}

	if err := r.db.Create(&newConnection).Error; err != nil {
		return nil, fmt.Errorf("не удалось добавить услугу в корзину: %w", err)
	}

	return &newConnection, nil
}

func (r *Repository) UploadPackageImage(packageID uint, file io.Reader, fileHeader *multipart.FileHeader) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var pkg domain.ConnectWifiPackages
		ctx := context.Background() // Создаем контекст для операций MinIO

		if err := tx.First(&pkg, packageID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("услуга с ID %d не найдена", packageID)
			}
			return err
		}

		if pkg.ImageUrl != nil && *pkg.ImageUrl != "" {
			parsedURL, err := url.Parse(*pkg.ImageUrl)
			if err == nil {
				objectName := strings.TrimPrefix(parsedURL.Path, "/"+r.minioClient.GetBucketName()+"/")
				if err := r.minioClient.DeleteObject(ctx, objectName); err != nil {
					log.Printf("Не удалось удалить старый файл из MinIO: %s", err)
				}
			}
		}

		newFileName := uuid.NewString() + filepath.Ext(fileHeader.Filename)

		err := r.minioClient.UploadObject(
			ctx,
			newFileName,
			file,
			fileHeader.Size,
			fileHeader.Header.Get("Content-Type"),
		)
		if err != nil {
			return fmt.Errorf("ошибка загрузки файла в MinIO: %w", err)
		}

		newImageUrl := fmt.Sprintf("%s/%s/%s", r.minioPublicURL, r.minioClient.GetBucketName(), newFileName)
		return tx.Model(&pkg).Update("image_url", newImageUrl).Error
	})
}