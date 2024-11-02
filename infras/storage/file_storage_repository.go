package storage

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/todennus/file-service/usecase/dto"
	"github.com/todennus/x/xcrypto"
)

type FileStorageRepository struct {
	imageBucket     string
	temporaryBucket string

	minioClient *minio.Client
}

func NewFileStorageRepository(
	minioClient *minio.Client,
	imageBucket string,
	temporaryBucket string,
) *FileStorageRepository {
	return &FileStorageRepository{
		imageBucket:     imageBucket,
		temporaryBucket: temporaryBucket,
		minioClient:     minioClient,
	}
}

func (repo *FileStorageRepository) ImageURL(name string) string {
	url := repo.minioClient.EndpointURL()
	url.Path = fmt.Sprintf("%s/%s", repo.imageBucket, name)
	return url.String()
}

func (repo *FileStorageRepository) ImageExists(ctx context.Context, name string) (bool, error) {
	_, err := repo.minioClient.StatObject(ctx, repo.imageBucket, name, minio.StatObjectOptions{})
	if err != nil {
		if errResponse := minio.ToErrorResponse(err); errResponse.Code == "NoSuchKey" {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (repo *FileStorageRepository) ImageSave(ctx context.Context, name string, temporaryFileName string) error {
	dst := minio.CopyDestOptions{Bucket: repo.imageBucket, Object: name}
	src := minio.CopySrcOptions{Bucket: repo.temporaryBucket, Object: temporaryFileName}
	if _, err := repo.minioClient.CopyObject(ctx, dst, src); err != nil {
		return err
	}

	return nil
}

func (repo *FileStorageRepository) StoreTemporary(
	ctx context.Context,
	filename string,
	content io.Reader,
	contentSize int64,
	contentType string,
) (*dto.StorageUploadFileMetadata, error) {
	sha256Reader := xcrypto.NewHashReader(content, sha256.New())
	info, err := repo.minioClient.PutObject(
		ctx,
		repo.temporaryBucket,
		filename,
		sha256Reader,
		contentSize,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return nil, err
	}

	return &dto.StorageUploadFileMetadata{
		Hash: base64.RawURLEncoding.EncodeToString(sha256Reader.Sum()),
		Size: int(info.Size),
	}, nil
}

func (repo *FileStorageRepository) GetTemporary(
	ctx context.Context,
	filename string,
) (io.ReadCloser, *dto.StorageDownloadFileMetadata, error) {
	obj, err := repo.minioClient.GetObject(
		ctx,
		repo.temporaryBucket,
		filename,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, nil, err
	}

	stat, err := obj.Stat()
	if err != nil {
		return nil, nil, err
	}

	return obj, &dto.StorageDownloadFileMetadata{
		Size: int(stat.Size),
		Type: stat.ContentType,
		Hash: stat.ChecksumSHA256,
	}, nil
}

func (repo *FileStorageRepository) RemoveTemporary(ctx context.Context, filename string) error {
	return repo.minioClient.RemoveObject(ctx, repo.temporaryBucket, filename, minio.RemoveObjectOptions{})
}
