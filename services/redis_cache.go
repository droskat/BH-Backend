package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/droskat/BH-Backend/models"
	"github.com/redis/go-redis/v9"
)

type RedisCacheService struct {
	client *redis.Client
}

func NewRedisCacheService(client *redis.Client) *RedisCacheService {
	return &RedisCacheService{client: client}
}

func imageKey(id uuid.UUID) string {
	return fmt.Sprintf("img:%s", id.String())
}

func userImagesKey(id uuid.UUID) string {
	return fmt.Sprintf("user_imgs:%s", id.String())
}

func (s *RedisCacheService) GetImage(ctx context.Context, imageID uuid.UUID) (*models.ImageByID, error) {
	data, err := s.client.Get(ctx, imageKey(imageID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("redis GetImage: %w", err)
	}

	var img models.ImageByID
	if err := json.Unmarshal(data, &img); err != nil {
		return nil, fmt.Errorf("redis GetImage unmarshal: %w", err)
	}
	return &img, nil
}

func (s *RedisCacheService) SetImage(ctx context.Context, imageID uuid.UUID, img *models.ImageByID, ttl time.Duration) error {
	data, err := json.Marshal(img)
	if err != nil {
		return fmt.Errorf("redis SetImage marshal: %w", err)
	}
	return s.client.Set(ctx, imageKey(imageID), data, ttl).Err()
}

func (s *RedisCacheService) DeleteImage(ctx context.Context, imageID uuid.UUID) error {
	return s.client.Del(ctx, imageKey(imageID)).Err()
}

func (s *RedisCacheService) GetUserImages(ctx context.Context, userID uuid.UUID) ([]models.ImageByUser, error) {
	data, err := s.client.Get(ctx, userImagesKey(userID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("redis GetUserImages: %w", err)
	}

	var images []models.ImageByUser
	if err := json.Unmarshal(data, &images); err != nil {
		return nil, fmt.Errorf("redis GetUserImages unmarshal: %w", err)
	}
	return images, nil
}

func (s *RedisCacheService) SetUserImages(ctx context.Context, userID uuid.UUID, images []models.ImageByUser, ttl time.Duration) error {
	data, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("redis SetUserImages marshal: %w", err)
	}
	return s.client.Set(ctx, userImagesKey(userID), data, ttl).Err()
}

func (s *RedisCacheService) DeleteUserImages(ctx context.Context, userID uuid.UUID) error {
	return s.client.Del(ctx, userImagesKey(userID)).Err()
}

func (s *RedisCacheService) PipelineSetPending(ctx context.Context, imageID uuid.UUID, userID uuid.UUID) error {
	pipe := s.client.Pipeline()

	pendingData, _ := json.Marshal(map[string]string{
		"image_id": imageID.String(),
		"status":   models.StatusPending,
	})
	pipe.Set(ctx, imageKey(imageID), pendingData, 48*time.Hour)
	pipe.Del(ctx, userImagesKey(userID))

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("[REDIS WARNING] pipeline set pending failed: %v", err)
		return err
	}
	return nil
}
