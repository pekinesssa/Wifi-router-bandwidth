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


type Order struct { // вот наша новая структура 
  ID    int // поля структур, которые передаются в шаблон
  Title string // ОБЯЗАТЕЛЬНО должны быть написаны с заглавной буквы (то есть публичными)
  Description string
  Price string
  Image string
}

type Cart struct { // вот наша новая структура 
  ID    int // поля структур, которые передаются в шаблон
  Goods []int // ОБЯЗАТЕЛЬНО должны быть написаны с заглавной буквы (то есть публичными)
  Address string
  CountOrders int
  Formula string
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

// func (r *Repository) GetCart(id int) (Cart, error) {
// 	if id != 1 {																																																									
// 		return Cart{}, fmt.Errorf("корзина с id %d не найдена", id)
// 	}

// 	allOrders, err := r.GetPackages()
// 	if err != nil {
// 		return Cart{}, fmt.Errorf("не удалось получить список товаров для создания корзины: %w", err)
// 	}

// 	goodsIDs := []int{}
// 	firstOrder := allOrders[0]
// 	goodsIDs = append(goodsIDs, firstOrder.ID)

// 	secondOrder := allOrders[1]
// 	goodsIDs = append(goodsIDs, secondOrder.ID)

// 	cart := Cart{
// 		ID:          1,
// 		Goods:       goodsIDs, 
// 		Address:     "ул. Проспект мира, д. 5 с.7, кв. 579",
// 		CountOrders: len(goodsIDs), 
// 		Formula:     "500 Мбит/с",
// 	}

// 	return cart, nil
// }

// func (r *Repository) GetOrder(id int) (Order, error) {
// 	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй 
// 	orders, err := r.GetPackages()
// 	if err != nil {
// 		return Order{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
// 	}

// 	for _, order := range orders {
// 		if order.ID == id {
// 			return order, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
// 		}
// 	}
// 	return Order{}, fmt.Errorf("заказ не найден") // тут нужна кастомная ошибка, чтобы понимать на каком этапе возникла ошибка и что произошло
// }

// func (r *Repository) GetOrdersByTitle(title string) ([]Order, error) {
// 	orders, err := r.GetPackages()
// 	if err != nil {
// 		return []Order{}, err
// 	}

// 	var result []Order
// 	for _, order := range orders {
// 		if strings.Contains(strings.ToLower(order.Title), strings.ToLower(title)) {
// 			result = append(result, order)
// 		}
// 	}

// 	return result, nil
// }