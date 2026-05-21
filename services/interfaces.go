package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/droskat/BH-Backend/models"
)

type ImageRepository interface {
	GetByID(ctx context.Context, imageID uuid.UUID) (*models.ImageByID, error)
	GetByUser(ctx context.Context, userID uuid.UUID) ([]models.ImageByUser, error)
	InsertImageByID(ctx context.Context, img *models.ImageByID) error
	InsertImageByUser(ctx context.Context, img *models.ImageByUser) error
	UpdateStatus(ctx context.Context, imageID, userID uuid.UUID, uploadDate time.Time, status string, width, height int, fileSize int64) error
	UpdateFilename(ctx context.Context, imageID uuid.UUID, filename string) error
	DeleteImage(ctx context.Context, imageID, userID uuid.UUID, uploadDate time.Time) error
}

type CacheService interface {
	GetImage(ctx context.Context, imageID uuid.UUID) (*models.ImageByID, error)
	SetImage(ctx context.Context, imageID uuid.UUID, img *models.ImageByID, ttl time.Duration) error
	DeleteImage(ctx context.Context, imageID uuid.UUID) error
	GetUserImages(ctx context.Context, userID uuid.UUID) ([]models.ImageByUser, error)
	SetUserImages(ctx context.Context, userID uuid.UUID, images []models.ImageByUser, ttl time.Duration) error
	DeleteUserImages(ctx context.Context, userID uuid.UUID) error
	PipelineSetPending(ctx context.Context, imageID uuid.UUID, userID uuid.UUID) error
}

type EventPublisher interface {
	PublishBatch(ctx context.Context, events []models.ImageEvent) error
	Close() error
}
