package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/paincake00/geocore/internal/config"
	delivery "github.com/paincake00/geocore/internal/delivery/http"
	"github.com/paincake00/geocore/internal/infrastructure/postgres"
	"github.com/paincake00/geocore/internal/infrastructure/redis"
	"github.com/paincake00/geocore/internal/usecase"
	"github.com/paincake00/geocore/internal/worker"
)

func main() {
	// Загружаем .env (опционально)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or failed to load, relying on environment variables")
	}

	// 1. Загрузка конфигурации
	cfg := config.Load()

	// 2. Подключение к базе данных (PostgreSQL)
	pgRepo, err := postgres.New(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer pgRepo.Close()

	// 3. Подключение к Redis
	redisRepo, err := redis.New(cfg.RedisAddr())
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redisRepo.Close()

	// 4. Инициализация сервисов (Application Layer)
	incidentService := usecase.NewIncidentService(pgRepo, redisRepo)
	// GeoService использует репозиторий инцидентов (postgres), репозиторий проверок (postgres), очередь (redis) и кеш (redis).
	// Обратите внимание: pgRepo реализует и IncidentRepository, и LocationCheckRepository.
	geoService := usecase.NewGeoService(pgRepo, pgRepo, redisRepo, redisRepo)

	// 5. Запуск воркера (Background Worker)
	w := worker.New(redisRepo, cfg.MockServerURL())
	workerCtx, workerCancel := context.WithCancel(context.Background())
	go w.Start(workerCtx)

	// 6. Инициализация HTTP-обработчика и роутера
	// Внедряем репозитории как "Pingers" для health-check
	handler := delivery.NewHandler(incidentService, geoService, pgRepo, redisRepo, cfg.APIKey())
	router := handler.InitRoutes()

	// 7. Запуск HTTP-сервера
	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort(),
		Handler: router,
	}

	go func() {
		log.Printf("Server listening on %s", cfg.HTTPPort())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 8. Graceful Shutdown (Плавное завершение)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	workerCancel() // Останавливаем воркер

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
