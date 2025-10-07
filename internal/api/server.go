package api

import (
	"Lab1/internal/app/handler"
	"Lab1/internal/app/repository"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория", err)
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()
	// добавляем наш html/шаблон
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/", handler.GetPackages)
	r.GET("/order/:id", handler.GetPackage) // вот наш новый обработчик
	r.GET("/cart", handler.GetEstimate) // вот наш новый обработчик

	r.POST("/cart/add/:service_id", handler.AddPackageeToEstimate)
	r.POST("/cart/delete/:id", handler.DeleteEstimate) // вот наш новый обработчик

	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", nil)
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	log.Println("Server down")
}

