package main

import (
	"Wi-Fi-router-bandwidth-backend/internal/api"
	"log"

	_ "Wi-Fi-router-bandwidth-backend/docs"
)

// @title DIA API
// @version 1.0
// @description API для расчета земляных работ при разработке котлована

// @host localhost:8080
// @BasePath /api
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите JWT токен в формате: "Bearer {your_token}".

func main() {
	log.Println("Application start!")
	api.StartServer()

	log.Println("Application terminated!")
}
