package models

import (
	"time"

	"github.com/google/uuid"
)

type ImageByID struct {
	ImageID          uuid.UUID `json:"image_id"`
	UserID           uuid.UUID `json:"user_id"`
	OriginalFilename string    `json:"original_filename"`
	UploadDate       time.Time `json:"upload_date"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	FileSize         int64     `json:"file_size"`
	FileType         string    `json:"file_type"`
	Status           string    `json:"status"`
}

type ImageByUser struct {
	UserID           uuid.UUID `json:"user_id"`
	UploadDate       time.Time `json:"upload_date"`
	ImageID          uuid.UUID `json:"image_id"`
	OriginalFilename string    `json:"original_filename"`
	FileSize         int64     `json:"file_size"`
	FileType         string    `json:"file_type"`
	Status           string    `json:"status"`
}
