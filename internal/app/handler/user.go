package handler

import (
	"Wi-Fi-router-bandwidth-backend/internal/app/middleware"
	"Wi-Fi-router-bandwidth-backend/internal/app/repository"
	"Wi-Fi-router-bandwidth-backend/internal/pkg"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

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
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

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
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
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
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	id, err := h.Repository.LoginUser(input)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Вход успешно выполнен": id})
}

func (h *Handler) PostLoginUserJWT(ctx *gin.Context) {
	var input repository.CreateUser
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.Repository.LoginUserJWT(input)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := pkg.GenerateToken(user.ID, user.Login, user.IsModerator)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "FНе получилось сгенерировать токен"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Аутентификация прошла успшено",
		"token":        token,
		"user_id":      user.ID,
		"is_moderator": user.IsModerator,
		"login":        user.Login,
	})
}


func (h *Handler) LogoutUser(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Авторизируйтесь!"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error:"не удалось"})
		return
	}

	tokenString := parts[1]
	claims, err := pkg.ValidateToken(tokenString)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный токен"})
		return
	}

	if h.Redis == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": nil})
		return
	}

	if claims.ExpiresAt == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "время действия токена не установлено"})
		return
	}

	expirationTime := claims.ExpiresAt.Time
	now := time.Now()
	ttl := expirationTime.Sub(now)

	if ttl <= 0 {
		logrus.Info("Действие токена уже закончилось, скип")
		ctx.JSON(http.StatusOK, gin.H{"message": "Время действия токена уже закончилось"})
		return
	}

	err = h.Redis.AddToBlacklist(ctx.Request.Context(), tokenString, ttl)
	if err != nil {
		logrus.Errorf("Не удалось добавить токен в блэклист: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось выйти"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Успешно удалось выйти"})
}
