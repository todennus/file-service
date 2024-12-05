package storage

import (
	"context"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/todennus/file-service/domain"
)

type FileStorageRepository struct {
	minioClient *minio.Client
}

func NewFileStorageRepository(
	minioClient *minio.Client,
) *FileStorageRepository {
	return &FileStorageRepository{
		minioClient: minioClient,
	}
}

func (repo *FileStorageRepository) Presign(
	ctx context.Context,
	file *domain.FileInfo,
	expiration time.Duration,
) (string, error) {
	bucket, filepath, found := strings.Cut(file.Metadata.Bucket, "/")
	filename := file.ID
	if found {
		filename = path.Join(filepath, filename)
	}

	url, err := repo.minioClient.PresignedGetObject(ctx, bucket, filename, expiration, url.Values{})
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (repo *FileStorageRepository) Store(
	ctx context.Context,
	file *domain.FileInfo,
	content io.Reader,
) error {
	bucket, filepath, found := strings.Cut(file.Metadata.Bucket, "/")
	filename := file.ID
	if found {
		filename = path.Join(filepath, filename)
	}

	size := int64(file.Metadata.Size)
	options := minio.PutObjectOptions{ContentType: file.Metadata.Type}

	if _, err := repo.minioClient.PutObject(ctx, bucket, filename, content, size, options); err != nil {
		return err
	}

	return nil
}
