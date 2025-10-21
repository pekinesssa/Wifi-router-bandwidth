package api

import (
	"Lab1/internal/app/handler"
	"Lab1/internal/app/repository"
	"Lab1/internal/pkg"
	"log"
	"os"

	// "net/http"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	minioCli, err := pkg.FromEnv()
	if err != nil {
		log.Fatalln("Ошибка инициализации клиента MinIO:", err)
	}

	minioPublicURL := os.Getenv("MINIO_PUBLIC_URL")
    if minioPublicURL == "" {
        minioPublicURL = "http://localhost:9000"
    }
	repo, err := repository.NewRepository(minioCli, minioPublicURL)
	if err != nil {
		logrus.Error("ошибка инициализации репозитория", err)
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()
	v1 := r.Group("/api")
	{
		v1.GET("/", handler.GetPackages) // получить все услуги (с фильтрацией)
		v1.GET("/package/:id", handler.GetPackage) // получить одну услугу
		// v1.GET("/estimate", handler.GetEstimate) // получить корзину 
		v1.POST("/package/add", handler.PostPackage) 
		v1.PUT("/package/edit/:id", handler.PutPackage)
		v1.DELETE("/package/delete/:id", handler.DeletePackage)
		v1.POST("/user/registrate", handler.PostUser) 
		v1.GET("/user/:id", handler.GetUser) 
		v1.PUT("/user/edit/:id", handler.PutUser) 
		v1.POST("/user/authenticate", handler.PostLoginUser) 
		v1.GET("/estimate/:id", handler.GetEstimate) 
		v1.PUT("/estimate/edit/:id", handler.PutEstimate)
		v1.GET("/estimate/fields/:id", handler.GetFieldEstimate) 
		v1.GET("/estimate", handler.GetListEstimate) 
		v1.PUT("/estimate/edit/data/:id", handler.PutCreatorEstimate)
		v1.PUT("/estimate/:id/moderate", handler.ModerateEstimate)
		v1.DELETE("/estimate/:id", handler.DeleteEstimate)
		v1.POST("/packages/:id/image", handler.UploadPackageImage)
		estimatePackages := v1.Group("/estimates/:estimate_id/packages")
		{
			estimatePackages.PUT("/:package_id", handler.UpdatePackageInEstimate)
			estimatePackages.DELETE("/:package_id", handler.DeletePackageFromEstimate)
		}
	}
	// добавляем наш html/шаблон
	// r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")




	r.POST("/estimate/add/:package_id", handler.AddPackageeToEstimate)
	// r.POST("/estimate/delete/:id", handler.DeleteEstimate) // вот наш новый обработчик

	// r.NoRoute(func(c *gin.Context) {
	// 	c.HTML(http.StatusNotFound, "404.html", nil)
	// })

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	log.Println("Server down")
}

