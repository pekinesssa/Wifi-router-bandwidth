package handler

import (
	"Lab1/internal/app/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
  Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
  return &Handler{
    Repository: r,
  }
}

func (h *Handler) GetOrders(ctx *gin.Context) {
	var orders []repository.Order
	var err error

	searchQuery := ctx.Query("query") // получаем значение из поля поиска
	if searchQuery == "" {            // если поле поиска пусто, то просто получаем из репозитория все записи
		orders, err = h.Repository.GetOrders()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		orders, err = h.Repository.GetOrdersByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}
	
	cart, err := h.Repository.GetCart(1)
	if err != nil {
		logrus.Warnf("Не удалось получить корзину для хедера: %v", err)
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":   time.Now().Format("15:04:05"),
		"orders": orders,
		"query":  searchQuery,
		"cart": cart, 
	})
}

func (h *Handler) GetOrder(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /order/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Error(err)
	}

	order, err := h.Repository.GetOrder(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "order.html", gin.H{
		"order": order,
	})
}

func (h *Handler) GetCart(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Errorf("неверный ID корзины: %v", err)
		return
	}

	cart, err := h.Repository.GetCart(id)
	if err != nil {
		logrus.Errorf("ошибка получения корзины: %v", err)
		return
	}

	var goodsDetails []repository.Order

	for _, goodID := range cart.Goods {
		order, err := h.Repository.GetOrder(goodID)
		if err != nil {
			logrus.Warnf("Товар с ID %d в корзине, но не найден в базе: %v", goodID, err)
			continue 
		}
		goodsDetails = append(goodsDetails, order)
	}

	ctx.HTML(http.StatusOK, "cart.html", gin.H{ 
		"cart":  cart,
		"goods": goodsDetails,
	})
}
