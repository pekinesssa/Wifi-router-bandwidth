package handler

import (
	"Wi-Fi-router-bandwidth-backend/internal/app/middleware"
	"Wi-Fi-router-bandwidth-backend/internal/app/repository"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ErrorResponse struct {
	Error string `json:"error" example:"Сообщение об ошибке"`
}

type PackageResponse struct {
	ID	uint   
	Title string
	ShortDescription string
	Description string
	Status string
	Price float64
	DeviceCount int            
	ImageURL	string  
}

type EstimateResponse struct {
	ID	uint           
	Status	string            
	CreatorID	uint              
	Packages	[]PackageResponse 
}

type EstimateListResponse struct {
	ID	uint      
	Status	string   
	CreatedAt  	time.Time
	CreatorID	uint      
	ModeratorID *int64   
}

func (h *Handler) GetEstimate(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// const currentUserID = 3
	idStr := ctx.Param("id") // указываем айди пользователя 
	cleanIdStr:=strings.TrimPrefix(idStr, ":")
	
	id, err := strconv.Atoi(cleanIdStr)
	if err != nil {
		logrus.Errorf("Ошибка преобразования id: %v", err)
		return
	}

	estimate, err := h.Repository.GetDraftEstimate(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusOK, estimate.ID)
			return
		}
		logrus.Errorf("ошибка получения корзины: %v", err)
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	CountOfPackages:= len(estimate.Bandwidthconnections)

	ctx.JSON(http.StatusOK, gin.H{"id заявки": estimate.ID, "id создателя заявки": estimate.CreatorID, "CountOfPackages": CountOfPackages, "status": estimate.Status})
}

func (h *Handler) PutEstimate(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var update repository.BandwidthAdded
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

	if err := h.Repository.PutEstimate(uint(id), update); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить данные"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Данные заявки успешно обновлены"})
}


func (h *Handler) GetFieldEstimate(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idStr := ctx.Param("id") 
	cleanIdStr:=strings.TrimPrefix(idStr, ":")
	
	id, err := strconv.Atoi(cleanIdStr)
	if err != nil {
		logrus.Errorf("Ошибка преобразования id: %v", err)
		return
	}

	estimate, err := h.Repository.GetFieldEstimate(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена"})
			return
		}
		logrus.Errorf("ошибка получения заявки: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	response := EstimateResponse{
		ID:        estimate.ID,
		Status:    estimate.Status,
		CreatorID: estimate.CreatorID,
		Packages:  make([]PackageResponse, 0, len(estimate.Bandwidthconnections)),
	}

	for _, conn := range estimate.Bandwidthconnections {
		var imageURL string 
   		if &conn.Connection != nil && conn.Connection.ImageUrl != nil {	
        	imageURL = *conn.Connection.ImageUrl 
   		}

		pkg := PackageResponse{
			ID:	conn.ConnectWifiPackagesID,
			Title:	conn.Connection.Title,   
			ShortDescription: conn.Connection.ShortDescription,
			Description: conn.Connection.Description,
			Status: conn.Connection.Status,
			Price: conn.Connection.Price,
			DeviceCount: conn.DeviceCount,
			ImageURL:	imageURL, 
		}
		response.Packages = append(response.Packages, pkg)
	}

	ctx.JSON(http.StatusOK, response)
}

// GetEstimate godoc
// @Summary Получение всех сетевых пакетов
// @Description Получение списка всех сетевых пакетов с возможностью фильтрации по названию
// @Tags estimate
// @Accept json
// @Produce json
// @Param packageTitle query string false "Фильтр по названию материала"
// @Success 200 {array} []EstimateListResponse
// @Failure 500 {object} ErrorResponse
// @Router /estimate [get]
func (h *Handler) GetListEstimate(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}
	var filters repository.EstimateFilterOptions
	estimates, err := h.Repository.GetFilteredEstimates(filters)
	if err != nil {
		logrus.Errorf("ошибка получения списка заявок: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	response := make([]EstimateListResponse, 0, len(estimates))
	for _, e := range estimates {
		item := EstimateListResponse{
			ID:	e.ID,
			Status:	e.Status,
			CreatedAt: e.CreatedAt,
			CreatorID: e.CreatorID, 
		}
        if e.ModeratorID.Valid {
            item.ModeratorID = &e.ModeratorID.Int64
        }
		response = append(response, item)
	}

	ctx.JSON(http.StatusOK, response)
}

// CreateMaterial godoc
// @Summary Создание нового материала
// @Description Создание материала (только для модераторов)
// @Tags estimate
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} repository.CreatorBandwidthAdded
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /edit/data/ [post]
func (h *Handler) PutCreatorEstimate(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var update repository.CreatorBandwidthAdded
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

	if err := h.Repository.PutCreatorEstimate(uint(id), update); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить данные"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Данные заявки успешно обновлены"})
}

func (h *Handler) ModerateEstimate(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID заявки"})
		return
	}
	estimateID := uint(id)

	const currentModeratorID uint = 2 

	var input repository.ModerateEstimateInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверное тело запроса: " + err.Error()})
		return
	}

	if input.Status != "completed" && input.Status != "rejected" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "статус может быть только 'completed' или 'rejected'"})
		return
	}

	err = h.Repository.ModerateEstimate(estimateID, currentModeratorID, input.Status)
	if err != nil {
		if strings.Contains(err.Error(), "не найдена") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "уже была обработана") || strings.Contains(err.Error(), "нельзя модерировать") {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		
		logrus.Errorf("ошибка при модерации заявки: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Заявка %d успешно обновлена до статуса '%s'", estimateID, input.Status)})
}

func (h *Handler) DeleteEstimate(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID заявки"})
		return
	}
	estimateID := uint(id)

	err = h.Repository.DeleteEstimate(estimateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Заявка с ID %d не найдена", estimateID)})
			return
		}

		logrus.Errorf("ошибка при удалении заявки: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Заявка с ID %d успешно удалена", estimateID)})
}
