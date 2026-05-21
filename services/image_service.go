package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/droskat/BH-Backend/models"
)

type ImageService struct {
	repo      ImageRepository
	cache     CacheService
	publisher EventPublisher
}

func NewImageService(repo ImageRepository, cache CacheService, publisher EventPublisher) *ImageService {
	return &ImageService{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
	}
}

func (s *ImageService) BulkUpload(ctx context.Context, userID uuid.UUID, items []models.BulkImageItem) (*models.BulkUploadResponse, error) {
	now := time.Now().UTC()
	records := make([]models.BulkRecordResponse, 0, len(items))
	events := make([]models.ImageEvent, 0, len(items))

	for _, item := range items {
		imageID := uuid.New()
		uploadURL := fmt.Sprintf("https://s3.amazonaws.com/bucket/%s?X-Amz-Signature=mock_%s", imageID, uuid.New().String()[:8])

		records = append(records, models.BulkRecordResponse{
			ImageID:   imageID,
			UploadURL: uploadURL,
			Metadata: map[string]interface{}{
				"image_id": imageID.String(),
				"status":   models.StatusPending,
			},
		})

		events = append(events, models.ImageEvent{
			EventType:        models.EventImageUploaded,
			ImageID:          imageID,
			UserID:           userID,
			OriginalFilename: item.OriginalFilename,
			FileType:         item.FileType,
			Status:           models.StatusPending,
			Timestamp:        now,
		})

		go func(imgID uuid.UUID, uID uuid.UUID) {
			cacheCtx := context.Background()
			if err := s.cache.PipelineSetPending(cacheCtx, imgID, uID); err != nil {
				log.Printf("[CACHE WARNING] PipelineSetPending failed for %s: %v", imgID, err)
			}
		}(imageID, userID)
	}

	if err := s.publisher.PublishBatch(ctx, events); err != nil {
		log.Printf("[KAFKA WARNING] PublishBatch failed: %v", err)
		return nil, fmt.Errorf("failed to publish events: %w", err)
	}

	return &models.BulkUploadResponse{
		Status:  "Allocation established. Complete S3 binary uploads.",
		Records: records,
	}, nil
}

func (s *ImageService) GetImageByID(ctx context.Context, imageID, requestingUserID uuid.UUID) (*models.ImageDetailResponse, error) {
	cached, err := s.cache.GetImage(ctx, imageID)
	if err != nil {
		log.Printf("[CACHE WARNING] GetImage failed for %s: %v, falling back to Cassandra", imageID, err)
	}

	if cached != nil {
		if cached.UserID != requestingUserID {
			return nil, ErrForbidden
		}
		go func() {
			_ = s.cache.SetImage(context.Background(), imageID, cached, 48*time.Hour)
		}()
		return toDetailResponse(cached), nil
	}

	img, err := s.repo.GetByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("repository GetByID: %w", err)
	}
	if img == nil {
		return nil, ErrNotFound
	}

	if img.UserID != requestingUserID {
		return nil, ErrForbidden
	}

	go func() {
		jitter := time.Duration(rand.Intn(6)+1) * time.Hour
		ttl := 48*time.Hour + jitter
		if cacheErr := s.cache.SetImage(context.Background(), imageID, img, ttl); cacheErr != nil {
			log.Printf("[CACHE WARNING] SetImage failed for %s: %v", imageID, cacheErr)
		}
	}()

	return toDetailResponse(img), nil
}

func (s *ImageService) GetUserImages(ctx context.Context, userID uuid.UUID) ([]models.ImageListItem, error) {
	cached, err := s.cache.GetUserImages(ctx, userID)
	if err != nil {
		log.Printf("[CACHE WARNING] GetUserImages failed for %s: %v, falling back to Cassandra", userID, err)
	}

	if cached != nil {
		return toListResponse(cached), nil
	}

	images, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repository GetByUser: %w", err)
	}

	go func() {
		jitter := time.Duration(rand.Intn(6)+1) * time.Hour
		ttl := 48*time.Hour + jitter
		if cacheErr := s.cache.SetUserImages(context.Background(), userID, images, ttl); cacheErr != nil {
			log.Printf("[CACHE WARNING] SetUserImages failed for %s: %v", userID, cacheErr)
		}
	}()

	return toListResponse(images), nil
}

func (s *ImageService) GetDownloadURL(ctx context.Context, imageID, requestingUserID uuid.UUID) (*models.DownloadResponse, error) {
	img, err := s.repo.GetByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("repository GetByID: %w", err)
	}
	if img == nil {
		return nil, ErrNotFound
	}
	if img.UserID != requestingUserID {
		return nil, ErrForbidden
	}

	downloadURL := fmt.Sprintf("https://s3.amazonaws.com/bucket/private/%s?X-Amz-Expires=300&Signature=mock_%s", imageID, uuid.New().String()[:8])
	return &models.DownloadResponse{
		ImageID:     imageID,
		DownloadURL: downloadURL,
	}, nil
}

func (s *ImageService) UpdateImage(ctx context.Context, imageID, requestingUserID uuid.UUID, req models.UpdateImageRequest) error {
	img, err := s.repo.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("repository GetByID: %w", err)
	}
	if img == nil {
		return ErrNotFound
	}
	if img.UserID != requestingUserID {
		return ErrForbidden
	}

	if err := s.repo.UpdateFilename(ctx, imageID, req.OriginalFilename); err != nil {
		return fmt.Errorf("repository UpdateFilename: %w", err)
	}

	go func() {
		bgCtx := context.Background()
		_ = s.cache.DeleteImage(bgCtx, imageID)
		_ = s.cache.DeleteUserImages(bgCtx, requestingUserID)
	}()

	return nil
}

func (s *ImageService) DeleteImage(ctx context.Context, imageID, requestingUserID uuid.UUID) error {
	img, err := s.repo.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("repository GetByID: %w", err)
	}
	if img == nil {
		return ErrNotFound
	}
	if img.UserID != requestingUserID {
		return ErrForbidden
	}

	if err := s.repo.DeleteImage(ctx, imageID, img.UserID, img.UploadDate); err != nil {
		return fmt.Errorf("repository DeleteImage: %w", err)
	}

	go func() {
		bgCtx := context.Background()
		_ = s.cache.DeleteImage(bgCtx, imageID)
		_ = s.cache.DeleteUserImages(bgCtx, requestingUserID)
	}()

	return nil
}

func toDetailResponse(img *models.ImageByID) *models.ImageDetailResponse {
	return &models.ImageDetailResponse{
		ImageID:          img.ImageID,
		UserID:           img.UserID,
		OriginalFilename: img.OriginalFilename,
		UploadDate:       img.UploadDate.Format(time.RFC3339),
		Width:            img.Width,
		Height:           img.Height,
		FileSize:         img.FileSize,
		FileType:         img.FileType,
		Status:           img.Status,
	}
}

func toListResponse(images []models.ImageByUser) []models.ImageListItem {
	result := make([]models.ImageListItem, 0, len(images))
	for _, img := range images {
		result = append(result, models.ImageListItem{
			ImageID:          img.ImageID,
			OriginalFilename: img.OriginalFilename,
			UploadDate:       img.UploadDate.Format(time.RFC3339),
			Status:           img.Status,
			ThumbnailURL:     fmt.Sprintf("https://cdn.platform.com/thumbnails/%s", img.ImageID),
		})
	}
	return result
}
