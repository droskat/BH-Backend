package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/droskat/BH-Backend/models"
	"github.com/droskat/BH-Backend/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Repository ---

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetByID(ctx context.Context, imageID uuid.UUID) (*models.ImageByID, error) {
	args := m.Called(ctx, imageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ImageByID), args.Error(1)
}

func (m *MockRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]models.ImageByUser, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ImageByUser), args.Error(1)
}

func (m *MockRepository) InsertImageByID(ctx context.Context, img *models.ImageByID) error {
	args := m.Called(ctx, img)
	return args.Error(0)
}

func (m *MockRepository) InsertImageByUser(ctx context.Context, img *models.ImageByUser) error {
	args := m.Called(ctx, img)
	return args.Error(0)
}

func (m *MockRepository) UpdateStatus(ctx context.Context, imageID, userID uuid.UUID, uploadDate time.Time, status string, width, height int, fileSize int64) error {
	args := m.Called(ctx, imageID, userID, uploadDate, status, width, height, fileSize)
	return args.Error(0)
}

func (m *MockRepository) UpdateFilename(ctx context.Context, imageID uuid.UUID, filename string) error {
	args := m.Called(ctx, imageID, filename)
	return args.Error(0)
}

func (m *MockRepository) DeleteImage(ctx context.Context, imageID, userID uuid.UUID, uploadDate time.Time) error {
	args := m.Called(ctx, imageID, userID, uploadDate)
	return args.Error(0)
}

// --- Mock Cache ---

type MockCache struct {
	mock.Mock
}

func (m *MockCache) GetImage(ctx context.Context, imageID uuid.UUID) (*models.ImageByID, error) {
	args := m.Called(ctx, imageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ImageByID), args.Error(1)
}

func (m *MockCache) SetImage(ctx context.Context, imageID uuid.UUID, img *models.ImageByID, ttl time.Duration) error {
	args := m.Called(ctx, imageID, img, ttl)
	return args.Error(0)
}

func (m *MockCache) DeleteImage(ctx context.Context, imageID uuid.UUID) error {
	args := m.Called(ctx, imageID)
	return args.Error(0)
}

func (m *MockCache) GetUserImages(ctx context.Context, userID uuid.UUID) ([]models.ImageByUser, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ImageByUser), args.Error(1)
}

func (m *MockCache) SetUserImages(ctx context.Context, userID uuid.UUID, images []models.ImageByUser, ttl time.Duration) error {
	args := m.Called(ctx, userID, images, ttl)
	return args.Error(0)
}

func (m *MockCache) DeleteUserImages(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCache) PipelineSetPending(ctx context.Context, imageID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, imageID, userID)
	return args.Error(0)
}

// --- Mock Publisher ---

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) PublishBatch(ctx context.Context, events []models.ImageEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

// --- Test Helpers ---

func setupTestRouter(svc *services.ImageService) (*gin.Engine, *ImageController) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := NewImageController(svc)
	return router, ctrl
}

func setUserID(c *gin.Context, userID uuid.UUID) {
	c.Set("user_id", userID)
}

// --- Tests ---

func TestBulkUpload_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockPub := new(MockPublisher)

	mockPub.On("PublishBatch", mock.Anything, mock.Anything).Return(nil)
	mockCache.On("PipelineSetPending", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewImageService(mockRepo, mockCache, mockPub)
	router, ctrl := setupTestRouter(svc)

	userID := uuid.New()
	router.POST("/api/v1/images/bulk", func(c *gin.Context) {
		setUserID(c, userID)
		ctrl.BulkUpload(c)
	})

	payload := models.BulkUploadRequest{
		Images: []models.BulkImageItem{
			{OriginalFilename: "test.png", FileType: "image/png"},
			{OriginalFilename: "photo.jpg", FileType: "image/jpeg"},
		},
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/bulk", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp models.BulkUploadResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(resp.Records))
	assert.Equal(t, "Allocation established. Complete S3 binary uploads.", resp.Status)

	mockPub.AssertExpectations(t)
}

func TestBulkUpload_ExceedsBatchLimit(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockPub := new(MockPublisher)

	svc := services.NewImageService(mockRepo, mockCache, mockPub)
	router, ctrl := setupTestRouter(svc)

	userID := uuid.New()
	router.POST("/api/v1/images/bulk", func(c *gin.Context) {
		setUserID(c, userID)
		ctrl.BulkUpload(c)
	})

	items := make([]models.BulkImageItem, 51)
	for i := range items {
		items[i] = models.BulkImageItem{OriginalFilename: "test.png", FileType: "image/png"}
	}
	payload := models.BulkUploadRequest{Images: items}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/bulk", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBulkUpload_InvalidFileType(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockPub := new(MockPublisher)

	svc := services.NewImageService(mockRepo, mockCache, mockPub)
	router, ctrl := setupTestRouter(svc)

	userID := uuid.New()
	router.POST("/api/v1/images/bulk", func(c *gin.Context) {
		setUserID(c, userID)
		ctrl.BulkUpload(c)
	})

	payload := models.BulkUploadRequest{
		Images: []models.BulkImageItem{
			{OriginalFilename: "test.exe", FileType: "application/exe"},
		},
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/bulk", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetImage_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockPub := new(MockPublisher)

	imageID := uuid.New()
	userID := uuid.New()

	mockCache.On("GetImage", mock.Anything, imageID).Return(nil, nil)
	mockRepo.On("GetByID", mock.Anything, imageID).Return(nil, nil)

	svc := services.NewImageService(mockRepo, mockCache, mockPub)
	router, ctrl := setupTestRouter(svc)

	router.GET("/api/v1/images/:id", func(c *gin.Context) {
		setUserID(c, userID)
		ctrl.GetImage(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images/"+imageID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetImage_Forbidden(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockPub := new(MockPublisher)

	imageID := uuid.New()
	ownerID := uuid.New()
	requestorID := uuid.New()

	mockCache.On("GetImage", mock.Anything, imageID).Return(nil, nil)
	mockRepo.On("GetByID", mock.Anything, imageID).Return(&models.ImageByID{
		ImageID:          imageID,
		UserID:           ownerID,
		OriginalFilename: "test.png",
		UploadDate:       time.Now(),
		Status:           "COMPLETED",
	}, nil)

	svc := services.NewImageService(mockRepo, mockCache, mockPub)
	router, ctrl := setupTestRouter(svc)

	router.GET("/api/v1/images/:id", func(c *gin.Context) {
		setUserID(c, requestorID)
		ctrl.GetImage(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images/"+imageID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetImage_Success_CacheHit(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockPub := new(MockPublisher)

	imageID := uuid.New()
	userID := uuid.New()

	cachedImg := &models.ImageByID{
		ImageID:          imageID,
		UserID:           userID,
		OriginalFilename: "cached.png",
		UploadDate:       time.Now(),
		Width:            1920,
		Height:           1080,
		FileSize:         204857,
		FileType:         "image/png",
		Status:           "COMPLETED",
	}

	mockCache.On("GetImage", mock.Anything, imageID).Return(cachedImg, nil)
	mockCache.On("SetImage", mock.Anything, imageID, cachedImg, mock.Anything).Return(nil)

	svc := services.NewImageService(mockRepo, mockCache, mockPub)
	router, ctrl := setupTestRouter(svc)

	router.GET("/api/v1/images/:id", func(c *gin.Context) {
		setUserID(c, userID)
		ctrl.GetImage(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images/"+imageID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ImageDetailResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, imageID, resp.ImageID)
	assert.Equal(t, "cached.png", resp.OriginalFilename)
}

func TestListImages_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockPub := new(MockPublisher)

	userID := uuid.New()

	mockCache.On("GetUserImages", mock.Anything, userID).Return(nil, nil)
	mockRepo.On("GetByUser", mock.Anything, userID).Return([]models.ImageByUser{
		{
			UserID:           userID,
			UploadDate:       time.Now(),
			ImageID:          uuid.New(),
			OriginalFilename: "img1.png",
			FileSize:         1024,
			FileType:         "image/png",
			Status:           "COMPLETED",
		},
	}, nil)
	mockCache.On("SetUserImages", mock.Anything, userID, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewImageService(mockRepo, mockCache, mockPub)
	router, ctrl := setupTestRouter(svc)

	router.GET("/api/v1/images", func(c *gin.Context) {
		setUserID(c, userID)
		ctrl.ListImages(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []models.ImageListItem
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resp))
}

func TestDeleteImage_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockPub := new(MockPublisher)

	imageID := uuid.New()
	userID := uuid.New()
	uploadDate := time.Now()

	mockCache.On("GetImage", mock.Anything, imageID).Return(nil, nil)
	mockRepo.On("GetByID", mock.Anything, imageID).Return(&models.ImageByID{
		ImageID:    imageID,
		UserID:     userID,
		UploadDate: uploadDate,
		Status:     "COMPLETED",
	}, nil)
	mockRepo.On("DeleteImage", mock.Anything, imageID, userID, uploadDate).Return(nil)
	mockCache.On("DeleteImage", mock.Anything, imageID).Return(nil)
	mockCache.On("DeleteUserImages", mock.Anything, userID).Return(nil)

	svc := services.NewImageService(mockRepo, mockCache, mockPub)
	router, ctrl := setupTestRouter(svc)

	router.DELETE("/api/v1/images/:id", func(c *gin.Context) {
		setUserID(c, userID)
		ctrl.DeleteImage(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/images/"+imageID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestGetImage_InvalidUUID(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	mockPub := new(MockPublisher)

	svc := services.NewImageService(mockRepo, mockCache, mockPub)
	router, ctrl := setupTestRouter(svc)

	userID := uuid.New()
	router.GET("/api/v1/images/:id", func(c *gin.Context) {
		setUserID(c, userID)
		ctrl.GetImage(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
