package models

import (
	"time"

	"github.com/google/uuid"
)

type ImageEvent struct {
	EventType        string    `json:"event_type"`
	ImageID          uuid.UUID `json:"image_id"`
	UserID           uuid.UUID `json:"user_id"`
	OriginalFilename string    `json:"original_filename"`
	FileType         string    `json:"file_type"`
	Status           string    `json:"status"`
	Timestamp        time.Time `json:"timestamp"`
}

const (
	StatusPending   = "PENDING"
	StatusCompleted = "COMPLETED"
	StatusFailed    = "FAILED"

	EventImageUploaded  = "IMAGE_UPLOADED"
	EventImageProcessed = "IMAGE_PROCESSED"
)
