package handler

import (
	"Lab1/internal/app/repository"
	"Lab1/internal/models"
	"net/http"
	"strconv"
	"strings"

	// "strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	// "golang.org/x/tools/go/packages"
	"gorm.io/gorm"
)

type Handler struct {
  Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
  return &Handler{
    Repository: r,
  }
}

func (h *Handler) GetPackages(ctx *gin.Context) {
	var orders []models.ConnectWifiPackages
	var err error
	const currentUserID = 1

	searchQuery := ctx.Query("query") // получаем значение из поля поиска
	if searchQuery == "" {            // если поле поиска пусто, то просто получаем из репозитория все записи
		orders, err = h.Repository.GetPackages()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		orders, err = h.Repository.GetPackagesByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}
	
	draftEstimate, err := h.Repository.GetDraftEstimate(currentUserID)
	var estimateCount int 
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Warnf("Не удалось получить корзину для хедера: %v", err)
	}
	if err == nil {
		estimateCount = len(draftEstimate.Bandwidthconnections)
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":   time.Now().Format("15:04:05"),
		"orders": orders,
		"query":  searchQuery,
		"estimateCount": estimateCount, 
	})
}

func (h *Handler) GetPackage(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /order/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше

	cleanIsStr := strings.TrimPrefix(idStr, ":")

	id, err := strconv.Atoi(cleanIsStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Errorf("Неверный формат ID: %v", err)
		ctx.String(http.StatusBadRequest, "Неверный формат ID")
		return
	}

	order, err := h.Repository.GetPackage(uint(id))
	if err != nil {
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "order.html", gin.H{
		"order": order,
	})
}

func (h *Handler) GetEstimate(ctx *gin.Context) {
	// idStr := ctx.Param("id")
	// id, err := strconv.Atoi(idStr)
	// if err != nil {
	// 	logrus.Errorf("неверный ID корзины: %v", err)
	// 	return
	// }

	const currentUserID = 1

	cart, err := h.Repository.GetDraftEstimate(currentUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.HTML(http.StatusOK, "cart.html", gin.H{
				"cart":  nil,
				"goods": nil,
			})
			return
		}
		logrus.Errorf("ошибка получения корзины: %v", err)
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	var totalPrice float64
	var goods []models.ConnectWifiPackages
	for _, conn := range cart.Bandwidthconnections {
		totalPrice += float64(conn.Connection.Price)
		goods = append(goods, conn.Connection)
	}

	// Заполняем расчетные поля для передачи в шаблон
	cart.TotalBandwidth = totalPrice

	ctx.HTML(http.StatusOK, "cart.html", gin.H{
		"cart":        cart,
		"goods":       goods,
		"countOrders": len(goods), // Передаем количество услуг
	})
}

func (h *Handler) AddPackageeToEstimate(ctx *gin.Context) {
	const currentUserID = 1

	serviceIdStr := ctx.Param("service_id")
	cleanServixeIsStr:= strings.TrimPrefix(serviceIdStr, ":")

	serviceID, err := strconv.Atoi(cleanServixeIsStr)
	if err != nil {
		logrus.Errorf("Ошибка преобразования id: %v", err)
		return 
	}

	err = h.Repository.AddPackageToEstimate(currentUserID, uint(serviceID))
	if err != nil {
		logrus.Errorf("ошибка добавления услуги в корзину: %v", err)
		return 
	}

	ctx.Redirect(http.StatusFound, "/cart")
}

func (h *Handler) DeleteEstimate(ctx *gin.Context) {
	const currentUserID = 1 // Хардкодим ID пользователя

	requestIdStr := ctx.Param("id")
	requestID, _ := strconv.Atoi(requestIdStr)

	err := h.Repository.DeleteEstimate(uint(requestID), currentUserID)
	if err != nil {
		logrus.Errorf("ошибка удаления заявки: %v", err)
	}

	ctx.Redirect(http.StatusFound, "/")
}
