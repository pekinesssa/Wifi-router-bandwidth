package handler

import (
	"Lab1/internal/app/repository"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) PostUser(ctx *gin.Context) {
	var input repository.CreateUser
	err := ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdUser, err := h.Repository.RegistrationUser(input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdUser)
}

func (h *Handler) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	cleanIsStr := strings.TrimPrefix(idStr, ":")

	id, err := strconv.Atoi(cleanIsStr)
	if err != nil {
		logrus.Errorf("Неверный формат ID: %v", err)
		ctx.JSON(http.StatusBadRequest, "Неверный формат ID")
		return
	}

	user, err := h.Repository.GetInfoUser(uint(id))
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, "Такого ID не существует")
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (h *Handler) PutUser(ctx *gin.Context) {
	var update repository.CreateUser
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
	if err := h.Repository.PutUser(uint(id), update); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить данные"})
		return 
	}

	ctx.JSON(http.StatusCreated, "Данные о пользователе успешно обнволены")
}

func (h *Handler) PostLoginUser(ctx *gin.Context) {
	var input repository.CreateUser
	err := ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id,err := h.Repository.LoginUser(input)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Вход успешно выполнен": id})
}

func (h *Handler) PostExitUser(ctx *gin.Context) {
	var input repository.CreateUser
	err := ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id,err := h.Repository.LoginUser(input)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Вход успешно выполнен": id})
}