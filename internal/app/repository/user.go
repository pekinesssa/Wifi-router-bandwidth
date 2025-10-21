package repository

import (
	"Lab1/internal/app/domain"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CreateUser struct{
	Login string
	Password string
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func CheckHashPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (r *Repository) RegistrationUser(imp CreateUser) (*domain.User, error) {
	HashedPassword, err := HashPassword(imp.Password)
	if err != nil {
		fmt.Println("Хэширование не удалось ошибка: %s", err.Error())
	}
	newUser := &domain.User{
		Login: imp.Login,
		HashedPassword: HashedPassword,
		IsModerator: false,
		IsDeleted: false,
	}

	result := r.db.Create(newUser)
	if result.Error != nil {
		fmt.Println("Ошибка при записи")
		return nil, result.Error
	}	
	return newUser, result.Error
}

func (r *Repository) GetInfoUser(id uint) (domain.User, error) {
	var User domain.User
	result := r.db.First(&User, id)
	if result.Error != nil {
		fmt.Println("Ошибка пользователь не найден, error: %s", result.Error)
	}
	return User, result.Error 
}

func (r *Repository) PutUser(id uint, inp CreateUser) error{
	updatedData := make(map[string]interface{})
	if inp.Login != "" {
		updatedData["Login"] = inp.Login
	}
	if inp.Password != "" {
		hashedPassword, err := HashPassword(inp.Password)
		if err != nil {
			fmt.Println("Хэширование не удалось ошибка: %s", err.Error())
		}
		updatedData["HashedPassword"] = hashedPassword
	}
	
	result := r.db.Model(&domain.User{}).Where("id = ?", id).Updates(updatedData)
	if result.Error != nil {
		fmt.Println("Ошибка при обновлнеии")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("пакет с ID %d не найден", id) 
	}

	return nil
}

func (r *Repository) LoginUser(imp CreateUser) (uint, error) {
	HashedPassword, err := HashPassword(imp.Password)
	if err != nil {
		fmt.Println("Хэширование не удалось ошибка: %s", err.Error())
	}
	authUser := &domain.User{
		Login: imp.Login,
		HashedPassword: HashedPassword,
		IsModerator: false,
		IsDeleted: false,
	}

	result := r.db.Where("login = ?", imp.Login).First(&authUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound{
			fmt.Println("Пользователь с таким логином не найден:", imp.Login)
			return 0, fmt.Errorf("неверный логин или пароль")
		}
		fmt.Println("Ошибка при поиске пользователя:", result.Error)
		return 0, result.Error
	}
	
	if !CheckHashPassword(imp.Password, authUser.HashedPassword){
		return 0, fmt.Errorf("неверный логин или пароль")
	}
	return authUser.ID, result.Error
}

func (r *Repository) UnloginUser(imp CreateUser) (uint, error) {
	HashedPassword, err := HashPassword(imp.Password)
	if err != nil {
		fmt.Println("Хэширование не удалось ошибка: %s", err.Error())
	}
	authUser := &domain.User{
		Login: imp.Login,
		HashedPassword: HashedPassword,
		IsModerator: false,
		IsDeleted: false,
	}

	result := r.db.Where("login = ?", imp.Login).First(&authUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound{
			fmt.Println("Пользователь с таким логином не найден:", imp.Login)
			return 0, fmt.Errorf("неверный логин или пароль")
		}
		fmt.Println("Ошибка при поиске пользователя:", result.Error)
		return 0, result.Error
	}
	
	if !CheckHashPassword(imp.Password, authUser.HashedPassword){
		return 0, fmt.Errorf("неверный логин или пароль")
	}
	return authUser.ID, result.Error
}