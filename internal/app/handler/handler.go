package handler

import (
	"Wi-Fi-router-bandwidth-backend/internal/app/domain"
	"Wi-Fi-router-bandwidth-backend/internal/app/middleware"
	"Wi-Fi-router-bandwidth-backend/internal/app/repository"
	"Wi-Fi-router-bandwidth-backend/internal/pkg"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Handler struct {
	Repository *repository.Repository
	Redis	*pkg.RedisClient
}

func NewHandler(r *repository.Repository, redisClient *pkg.RedisClient) *Handler {
	return &Handler{
		Repository: r,
		Redis:	redisClient,
	}
}

const bucketName = "main" 

func (h *Handler) GetPackages(ctx *gin.Context) {
	var packages []domain.ConnectWifiPackages
	var err error
	const currentUserID = 1

	searchQuery := ctx.Query("query") // получаем значение из поля поиска
	if searchQuery == "" {            // если поле поиска пусто, то просто получаем из репозитория все записи
		packages, err = h.Repository.GetPackages()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		packages, err = h.Repository.GetPackagesByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}

	// draftEstimate, err := h.Repository.GetDraftEstimate(currentUserID)
	// var estimateCount int
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Warnf("Не удалось получить корзину для хедера: %v", err)
	}
	// if err == nil {
	// 	estimateCount = len(draftEstimate.Bandwidthconnections)
	// }

	// ctx.JSON(http.StatusOK, gin.H{
	// 	"time":   time.Now().Format("15:04:05"),
	// 	"packages": packages,
	// 	"query":  searchQuery,
	// 	"estimateCount": estimateCount,
	// })

	ctx.JSON(http.StatusOK, packages)
}

func (h *Handler) GetPackage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	cleanIsStr := strings.TrimPrefix(idStr, ":")

	id, err := strconv.Atoi(cleanIsStr)
	if err != nil {
		logrus.Errorf("Неверный формат ID: %v", err)
		ctx.JSON(http.StatusBadRequest, "Неверный формат ID")
		return
	}

	packages, err := h.Repository.GetPackage(uint(id))
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, "Такого ID не существует")
		return
	}

	ctx.JSON(http.StatusOK, packages)
}

func (h *Handler) PostPackage(ctx *gin.Context) {
	var input repository.CreatePackage
	err := ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdPackage, err := h.Repository.PostPackage(input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("Creted package with", input.Title)
	ctx.JSON(http.StatusCreated, createdPackage)
}

func (h *Handler) PutPackage(ctx *gin.Context) {
	var update repository.UpdatePackage
	idStr := ctx.Param("id")
	cleanIsStr := strings.TrimPrefix(idStr, ":")

	id, err := strconv.Atoi(cleanIsStr)
	if err != nil {
		logrus.Errorf("Неверный формат ID: %v", err)
		ctx.JSON(http.StatusBadRequest, "Неверный формат ID")
		return
	}

	err = ctx.ShouldBindJSON(&update)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repository.PutPackage(uint(id), update); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить данные"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Пакет успешно обновлен"})
}

func (h *Handler) DeletePackage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	cleanStr := strings.TrimPrefix(idStr, ":")

	id, err := strconv.Atoi(cleanStr)
	if err != nil {
		logrus.Errorf("Неверный формат ID: %v", err)
		ctx.JSON(http.StatusBadRequest, "Неверный формат ID")
		return
	}

	err = h.Repository.DeletePackage(uint(id))
	if err != nil {
		logrus.Errorf("Неверный формат ID: %v", err)
		ctx.JSON(http.StatusBadRequest, "Неверный формат ID")
		return
	}	

	ctx.JSON(http.StatusNoContent, "удалено")
}

func (h *Handler) AddPackageeToEstimate(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	const currentUserID = 3

	serviceIdStr := ctx.Param("package_id")
	// cleanServixeIsStr:= strings.TrimPrefix(serviceIdStr, ":")

	serviceID, err := strconv.Atoi(serviceIdStr)
	if err != nil {
		logrus.Errorf("Ошибка преобразования id: %v", err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(3)

	estimate, err := h.Repository.AddPackageToEstimate(currentUserID, uint(serviceID), randomNumber)
	if err != nil {
		logrus.Errorf("ошибка добавления услуги в корзину: %v", err)
		return
	}

	ctx.JSON(http.StatusCreated, estimate)
}

func (h *Handler) UploadPackageImage(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID услуги"})
		return
	}

	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "файл 'image' не найден в запросе"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть файл"})
		return
	}
	defer file.Close() 

	err = h.Repository.UploadPackageImage(uint(id), file, fileHeader)
	if err != nil {
		if strings.Contains(err.Error(), "не найдена") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		logrus.Errorf("ошибка при загрузке изображения: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Изображение успешно загружено и обновлено"})
}