package worker

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/droskat/BH-Backend/models"
	"github.com/droskat/BH-Backend/services"
	"github.com/segmentio/kafka-go"
)

type ImageProcessor struct {
	reader *kafka.Reader
	repo   services.ImageRepository
	cache  services.CacheService
}

func NewImageProcessor(reader *kafka.Reader, repo services.ImageRepository, cache services.CacheService) *ImageProcessor {
	return &ImageProcessor{
		reader: reader,
		repo:   repo,
		cache:  cache,
	}
}

func (p *ImageProcessor) Start(ctx context.Context) {
	log.Println("[WORKER] Image processor started")

	for {
		select {
		case <-ctx.Done():
			log.Println("[WORKER] Shutting down processor")
			return
		default:
			msg, err := p.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[WORKER ERROR] reading message: %v", err)
				continue
			}
			p.processMessage(ctx, msg)
		}
	}
}

func (p *ImageProcessor) processMessage(ctx context.Context, msg kafka.Message) {
	var event models.ImageEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("[WORKER ERROR] unmarshal event: %v", err)
		return
	}

	log.Printf("[WORKER] Processing image %s for user %s", event.ImageID, event.UserID)

	simulatedWidth := 800 + rand.Intn(3200)
	simulatedHeight := 600 + rand.Intn(2400)
	simulatedFileSize := int64(50000 + rand.Intn(500000))

	now := event.Timestamp
	imgByID := &models.ImageByID{
		ImageID:          event.ImageID,
		UserID:           event.UserID,
		OriginalFilename: event.OriginalFilename,
		UploadDate:       now,
		Width:            simulatedWidth,
		Height:           simulatedHeight,
		FileSize:         simulatedFileSize,
		FileType:         event.FileType,
		Status:           models.StatusPending,
	}

	if err := p.repo.InsertImageByID(ctx, imgByID); err != nil {
		log.Printf("[WORKER ERROR] InsertImageByID: %v", err)
		return
	}

	imgByUser := &models.ImageByUser{
		UserID:           event.UserID,
		UploadDate:       now,
		ImageID:          event.ImageID,
		OriginalFilename: event.OriginalFilename,
		FileSize:         simulatedFileSize,
		FileType:         event.FileType,
		Status:           models.StatusPending,
	}

	if err := p.repo.InsertImageByUser(ctx, imgByUser); err != nil {
		log.Printf("[WORKER ERROR] InsertImageByUser: %v", err)
		return
	}

	time.Sleep(time.Duration(100+rand.Intn(400)) * time.Millisecond)

	if err := p.repo.UpdateStatus(ctx, event.ImageID, event.UserID, now, models.StatusCompleted, simulatedWidth, simulatedHeight, simulatedFileSize); err != nil {
		log.Printf("[WORKER ERROR] UpdateStatus: %v", err)
		return
	}

	if err := p.cache.DeleteImage(ctx, event.ImageID); err != nil {
		log.Printf("[WORKER WARNING] cache DeleteImage: %v", err)
	}
	if err := p.cache.DeleteUserImages(ctx, event.UserID); err != nil {
		log.Printf("[WORKER WARNING] cache DeleteUserImages: %v", err)
	}

	log.Printf("[WORKER] Completed processing image %s - dimensions: %dx%d", event.ImageID, simulatedWidth, simulatedHeight)
}

func (p *ImageProcessor) Close() error {
	return p.reader.Close()
}
