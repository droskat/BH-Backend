package services

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/droskat/BH-Backend/models"
)

type CassandraRepository struct {
	session *gocql.Session
}

func NewCassandraRepository(session *gocql.Session) *CassandraRepository {
	return &CassandraRepository{session: session}
}

func (r *CassandraRepository) GetByID(ctx context.Context, imageID uuid.UUID) (*models.ImageByID, error) {
	var img models.ImageByID
	query := `SELECT image_id, user_id, original_filename, upload_date, width, height, file_size, file_type, status 
	           FROM images_by_id WHERE image_id = ?`

	err := r.session.Query(query, gocql.UUID(imageID)).WithContext(ctx).Scan(
		&img.ImageID, &img.UserID, &img.OriginalFilename,
		&img.UploadDate, &img.Width, &img.Height,
		&img.FileSize, &img.FileType, &img.Status,
	)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("cassandra GetByID: %w", err)
	}
	return &img, nil
}

func (r *CassandraRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]models.ImageByUser, error) {
	query := `SELECT user_id, upload_date, image_id, original_filename, file_size, file_type, status 
	           FROM images_by_user WHERE user_id = ?`

	iter := r.session.Query(query, gocql.UUID(userID)).WithContext(ctx).Iter()
	var images []models.ImageByUser

	var img models.ImageByUser
	for iter.Scan(&img.UserID, &img.UploadDate, &img.ImageID, &img.OriginalFilename, &img.FileSize, &img.FileType, &img.Status) {
		images = append(images, img)
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("cassandra GetByUser: %w", err)
	}
	return images, nil
}

func (r *CassandraRepository) InsertImageByID(ctx context.Context, img *models.ImageByID) error {
	query := `INSERT INTO images_by_id (image_id, user_id, original_filename, upload_date, width, height, file_size, file_type, status)
	           VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	return r.session.Query(query,
		gocql.UUID(img.ImageID), gocql.UUID(img.UserID), img.OriginalFilename,
		img.UploadDate, img.Width, img.Height,
		img.FileSize, img.FileType, img.Status,
	).WithContext(ctx).Exec()
}

func (r *CassandraRepository) InsertImageByUser(ctx context.Context, img *models.ImageByUser) error {
	query := `INSERT INTO images_by_user (user_id, upload_date, image_id, original_filename, file_size, file_type, status)
	           VALUES (?, ?, ?, ?, ?, ?, ?)`
	return r.session.Query(query,
		gocql.UUID(img.UserID), img.UploadDate, gocql.UUID(img.ImageID),
		img.OriginalFilename, img.FileSize, img.FileType, img.Status,
	).WithContext(ctx).Exec()
}

func (r *CassandraRepository) UpdateStatus(ctx context.Context, imageID, userID uuid.UUID, uploadDate time.Time, status string, width, height int, fileSize int64) error {
	q1 := `UPDATE images_by_id SET status = ?, width = ?, height = ?, file_size = ? WHERE image_id = ?`
	q2 := `UPDATE images_by_user SET status = ? WHERE user_id = ? AND upload_date = ?`

	if err := r.session.Query(q1, status, width, height, fileSize, gocql.UUID(imageID)).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("cassandra UpdateStatus images_by_id: %w", err)
	}
	if err := r.session.Query(q2, status, gocql.UUID(userID), uploadDate).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("cassandra UpdateStatus images_by_user: %w", err)
	}
	return nil
}

func (r *CassandraRepository) UpdateFilename(ctx context.Context, imageID uuid.UUID, filename string) error {
	query := `UPDATE images_by_id SET original_filename = ? WHERE image_id = ?`
	return r.session.Query(query, filename, gocql.UUID(imageID)).WithContext(ctx).Exec()
}

func (r *CassandraRepository) DeleteImage(ctx context.Context, imageID, userID uuid.UUID, uploadDate time.Time) error {
	q1 := `DELETE FROM images_by_id WHERE image_id = ?`
	q2 := `DELETE FROM images_by_user WHERE user_id = ? AND upload_date = ?`

	if err := r.session.Query(q1, gocql.UUID(imageID)).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("cassandra DeleteImage images_by_id: %w", err)
	}
	if err := r.session.Query(q2, gocql.UUID(userID), uploadDate).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("cassandra DeleteImage images_by_user: %w", err)
	}
	return nil
}
