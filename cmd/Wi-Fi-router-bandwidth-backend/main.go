package main

import (
	"Lab1/internal/api"
	"log"
    _ "Lab1/docs" 

)


// @title           API для лабораторной работы
// @version         1.0
// @description     Это сервер API с аутентификацией по JWT и управлением пользователями.

// @contact.name   Kaverina 
// @contact.url    https://t.me/ya_kaverina

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description "Для доступа к защищенным методам введите 'Bearer ' и после пробела ваш JWT токен."
func main(){
	log.Println("Application start!")
	api.StartServer()
	
	log.Println("Application terminated!")
}