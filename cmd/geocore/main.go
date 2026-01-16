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
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Database
	pgRepo, err := postgres.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer pgRepo.Close()

	// 3. Redis
	redisRepo, err := redis.New(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redisRepo.Close()

	// 4. Services
	incidentService := usecase.NewIncidentService(pgRepo, redisRepo)
	geoService := usecase.NewGeoService(pgRepo, pgRepo, redisRepo, redisRepo)

	// 5. Worker
	w := worker.New(redisRepo, cfg.MockServerURL)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	go w.Start(workerCtx)

	// 6. HTTP Handler
	// Inject repos as Pingers
	handler := delivery.NewHandler(incidentService, geoService, pgRepo, redisRepo)
	router := handler.InitRoutes()

	// 7. Server
	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	go func() {
		log.Printf("Server listening on %s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 8. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	workerCancel() // Stop worker

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
