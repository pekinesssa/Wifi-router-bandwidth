package api

import (
	"Wi-Fi-router-bandwidth-backend/internal/app/handler"
	"Wi-Fi-router-bandwidth-backend/internal/app/middleware"
	"Wi-Fi-router-bandwidth-backend/internal/app/repository"
	"Wi-Fi-router-bandwidth-backend/internal/pkg"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	redisClient, err := pkg.NewRedisClient("localhost:6379", "password", 0)
	if err != nil {
		redisClient = nil
	} else {
		defer redisClient.Close()
	}

	handler := handler.NewHandler(repo, redisClient)
	
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	public := r.Group("/api")
	{
		public.GET("/", handler.GetPackages)
		public.GET("/package/:id", handler.GetPackage)
		public.POST("/user/registrate", handler.PostUser) 
		// public.POST("/user/authenticate", handler.PostLoginUser)
		public.POST("/user/authenticate", handler.PostLoginUserJWT)
	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthRequired(redisClient))
	{
		protected.POST("/user/logout", handler.LogoutUser)
		protected.GET("/user/:id", handler.GetUser)
		protected.PUT("/user/edit/:id", handler.PutUser)
		protected.GET("/estimate/:id", handler.GetEstimate)
		protected.PUT("/estimate/edit/:id", handler.PutEstimate)
		protected.GET("/estimate/fields/:id", handler.GetFieldEstimate)
		protected.PUT("/estimate/edit/data/:id", handler.PutCreatorEstimate)
		protected.DELETE("/estimate/:id", handler.DeleteEstimate)
		protected.POST("/estimate/add/:package_id", handler.AddPackageeToEstimate)
		protected.POST("/packages/:id/image", handler.UploadPackageImage)
		
		estimatePackages := protected.Group("/estimates/:estimate_id/packages")
		{
			estimatePackages.PUT("/:package_id", handler.UpdatePackageInEstimate)
			estimatePackages.DELETE("/:package_id", handler.DeletePackageFromEstimate)
		}

		moderator := protected.Group("/")
		moderator.Use(middleware.ModeratorRequired())
		{
			moderator.GET("/estimate", handler.GetListEstimate)
			moderator.PUT("/estimate/:id/moderate", handler.ModerateEstimate)
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

type pingReq struct{}
type pingResp struct {
   Status string `json:"status"`
}

// Ping godoc
// @Summary      Show hello text
// @Description  very very friendly response
// @Tags         Tests
// @Produce      json
// @Success      200  {object}  pingResp
// @Router       /ping/{name} [get]
func Ping(gCtx *gin.Context) {
   name := gCtx.Param("name")
   gCtx.String(http.StatusOK, "Hello %s", name)
}
