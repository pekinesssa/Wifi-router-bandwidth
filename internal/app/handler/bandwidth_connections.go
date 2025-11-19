package handler

import (
	"errors"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UpdatePackageInput struct {
	DeviceCount int `json:"device_count" binding:"required,gte=1"`
}

func (h *Handler) UpdatePackageInEstimate(ctx *gin.Context) {
	estimateIDStr := ctx.Param("estimate_id")
	packageIDStr := ctx.Param("package_id")

	estimateID, err := strconv.Atoi(estimateIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID оценки сетевых пакетов"})
		return
	}
	packageID, err := strconv.Atoi(packageIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID сетевого пакета"})
		return
	}

	var input UpdatePackageInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверное тело запроса: " + err.Error()})
		return
	}

	err = h.Repository.UpdatePackageInEstimate(uint(estimateID), uint(packageID), input.DeviceCount)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "указанная услуга в данной оценки сетевых пакетов не найдена"})
			return
		}
		logrus.Errorf("ошибка при обновлении услуги в заявке: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Количество устройств для сетевого пакета успешно обновлено"})
}

func (h *Handler) DeletePackageFromEstimate(ctx *gin.Context) {
	estimateIDStr := ctx.Param("estimate_id")
	packageIDStr := ctx.Param("package_id")

	estimateID, err := strconv.Atoi(estimateIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID оценки сетевых пакетов"})
		return
	}
	packageID, err := strconv.Atoi(packageIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID сетевого пакета"})
		return
	}

	err = h.Repository.DeletePackageFromEstimate(uint(estimateID), uint(packageID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "указанный сетевой пакет в данной оценке сетевых пакетов не найдена"})
			return
		}
		logrus.Errorf("ошибка при удалении услуги из заявки: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Сетевой пакет успешно удалена из оценки сетевых пакетов"})
}