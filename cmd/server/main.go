package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/droskat/BH-Backend/config"
	"github.com/droskat/BH-Backend/controllers"
	"github.com/droskat/BH-Backend/connectors"
	"github.com/droskat/BH-Backend/internal/worker"
	"github.com/droskat/BH-Backend/middlewares"
	"github.com/droskat/BH-Backend/services"
)

func main() {
	cfg := config.Load()

	cassSession, err := connectors.NewCassandraSession(cfg.Cassandra)
	if err != nil {
		log.Fatalf("[FATAL] Cassandra connection failed: %v", err)
	}
	defer cassSession.Close()

	redisClient, err := connectors.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Printf("[WARNING] Redis connection failed: %v - running in degraded mode", err)
	}
	defer func() {
		if redisClient != nil {
			_ = redisClient.Close()
		}
	}()

	kafkaWriter := connectors.NewKafkaWriter(cfg.Kafka)
	kafkaReader := connectors.NewKafkaReader(cfg.Kafka, "image-processor-group")

	repo := services.NewCassandraRepository(cassSession)
	cache := services.NewRedisCacheService(redisClient)
	publisher := services.NewKafkaPublisher(kafkaWriter)

	imageService := services.NewImageService(repo, cache, publisher)

	authEngine := middlewares.NewAuthEngine(cfg.JWT)
	imageCtrl := controllers.NewImageController(imageService)
	authCtrl := controllers.NewAuthController(authEngine)

	router := setupRouter(authEngine, imageCtrl, authCtrl)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	processor := worker.NewImageProcessor(kafkaReader, repo, cache)
	go processor.Start(workerCtx)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[SERVER] Starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[SERVER] Shutting down gracefully...")

	workerCancel()
	_ = processor.Close()
	_ = publisher.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[FATAL] Server forced shutdown: %v", err)
	}

	log.Println("[SERVER] Exited cleanly")
}
