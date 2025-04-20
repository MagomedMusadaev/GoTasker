package http

import (
	"GoTasker/internal/handler/analytics"
	"GoTasker/internal/handler/auth"
	"GoTasker/internal/handler/tasks"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes настраивает маршруты
func SetupRoutes(r *gin.Engine,
	taskHandler *tasks.TaskHandler,
	analyticHandler *analytics.TaskAnalyticsHandler,
	authHandler *auth.UserAuthHandler,
) {
	taskGroup := r.Group("/tasks")
	{
		taskGroup.GET("", taskHandler.GetAll)        // Получение списка задач
		taskGroup.POST("", taskHandler.Create)       // Создание задачи
		taskGroup.PUT("/:id", taskHandler.Update)    // Обновление задачи
		taskGroup.DELETE("/:id", taskHandler.Delete) // Удаление задачи

		taskGroup.POST("/import", taskHandler.Import) // Импорт задач
		taskGroup.GET("/export", taskHandler.Export)  // Экспорт задач
	}

	analyticGroup := r.Group("/analytics")
	{
		analyticGroup.GET("", analyticHandler.GetAnalytics) // Получение аналитики
	}

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register) // Регистрация пользователя
		authGroup.POST("/login", authHandler.Login)       // Вход и получение JWT (access and refresh)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
