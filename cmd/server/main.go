package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"geo-alert-core/internal/config"
	"geo-alert-core/internal/handler"
	"geo-alert-core/internal/infrastructure/postgres"
	"geo-alert-core/internal/infrastructure/redis"
	"geo-alert-core/internal/infrastructure/webhook"
	"geo-alert-core/internal/middleware"
	"geo-alert-core/internal/repository"
	"geo-alert-core/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключаемся к PostgreSQL
	db, err := postgres.NewDB(cfg.GetPostgresDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Подключаемся к Redis
	redisClient, err := redis.NewClient(cfg.GetRedisAddr(), cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Создаем вебхук отправитель
	webhookSender := webhook.NewSender(
		cfg.WebhookURL,
		cfg.WebhookRetryAttempts,
		cfg.WebhookRetryDelaySec,
	)

	// Создаем репозитории
	incidentRepo := repository.NewPostgresIncidentRepository(db)
	locationCheckRepo := repository.NewPostgresLocationCheckRepository(db)

	// Создаем сервисы
	incidentService := service.NewIncidentService(incidentRepo)
	locationService := service.NewLocationService(
		incidentRepo,
		locationCheckRepo,
		redisClient.GetClient(),
		webhookSender,
	)
	statsService := service.NewStatsService(incidentRepo)

	// Связываем сервисы для инвалидации кэша
	incidentService.SetLocationService(locationService)

	// Создаем handlers
	healthHandler := handler.NewHealthHandler()
	incidentHandler := handler.NewIncidentHandler(incidentService)
	locationHandler := handler.NewLocationHandler(locationService)
	statsHandler := handler.NewStatsHandler(statsService)

	// Настраиваем роутер
	router := setupRouter(
		cfg.APIKey,
		healthHandler,
		incidentHandler,
		locationHandler,
		statsHandler,
	)

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter(
	apiKey string,
	healthHandler *handler.HealthHandler,
	incidentHandler *handler.IncidentHandler,
	locationHandler *handler.LocationHandler,
	statsHandler *handler.StatsHandler,
) *gin.Engine {
	router := gin.Default()

	// Публичные эндпоинты (без API key)
	public := router.Group("/api/v1")
	{
		public.GET("/system/health", healthHandler.Health)
		public.POST("/location/check", locationHandler.CheckLocation)
	}

	// Защищенные эндпоинты
	protected := router.Group("/api/v1")
	protected.Use(middleware.APIKeyAuth(apiKey))
	{
		// Управление инцидентами
		incidents := protected.Group("/incidents")
		{
			incidents.POST("", incidentHandler.Create)
			incidents.GET("", incidentHandler.GetAll)
			incidents.GET("/:id", incidentHandler.GetByID)
			incidents.PUT("/:id", incidentHandler.Update)
			incidents.DELETE("/:id", incidentHandler.Delete)
		}

		// Статистика
		protected.GET("/incidents/stats", statsHandler.GetStats)
	}

	return router
}
