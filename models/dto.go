package models

import "github.com/google/uuid"

// --- Request DTOs ---

type BulkImageItem struct {
	OriginalFilename string `json:"original_filename" binding:"required,min=1,max=255"`
	FileType         string `json:"file_type" binding:"required,oneof=image/png image/jpeg image/gif image/webp image/bmp image/tiff"`
}

type BulkUploadRequest struct {
	Images []BulkImageItem `json:"images" binding:"required,min=1,max=50,dive"`
}

type UpdateImageRequest struct {
	OriginalFilename string `json:"original_filename" binding:"required,min=1,max=255"`
}

type LoginRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// --- Response DTOs ---

type BulkRecordResponse struct {
	ImageID   uuid.UUID              `json:"image_id"`
	UploadURL string                 `json:"upload_url"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type BulkUploadResponse struct {
	Status  string               `json:"status"`
	Records []BulkRecordResponse `json:"records"`
}

type ImageListItem struct {
	ImageID          uuid.UUID `json:"image_id"`
	OriginalFilename string    `json:"original_filename"`
	UploadDate       string    `json:"upload_date"`
	Status           string    `json:"status"`
	ThumbnailURL     string    `json:"thumbnail_url"`
}

type ImageDetailResponse struct {
	ImageID          uuid.UUID `json:"image_id"`
	UserID           uuid.UUID `json:"user_id"`
	OriginalFilename string    `json:"original_filename"`
	UploadDate       string    `json:"upload_date"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	FileSize         int64     `json:"file_size"`
	FileType         string    `json:"file_type"`
	Status           string    `json:"status"`
}

type DownloadResponse struct {
	ImageID     uuid.UUID `json:"image_id"`
	DownloadURL string    `json:"download_url"`
}

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
