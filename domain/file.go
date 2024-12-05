package domain

import (
	"time"

	"github.com/todennus/x/mime"
	"github.com/todennus/x/xcrypto"
	"github.com/xybor-x/snowflake"
)

type UploadPolicy struct {
	Token string

	// AllowedTypes specifies the permitted content types for the uploaded file.
	// It defines which MIME types are acceptable for file uploads.
	AllowedTypes []string

	// Maxsize defines the maximum size, in bytes, of the uploaded file.
	MaxSize int64

	// UserID represents who can upload file.
	UserID snowflake.ID

	ExpiresAt time.Time
}

// FileMetadata contains additional information about a file.
type FileMetadata struct {
	// Bucket determines where this file is stored (it usually is the folder name).
	Bucket string

	// Type represents the MIME type of the file.
	Type string

	// Size is the size of the file content in bytes.
	Size int
}

type FileInfo struct {
	ID        string
	Metadata  *FileMetadata
	CreatedAt time.Time
}

type FileOwnership struct {
	ID       snowflake.ID
	FileID   string
	UserID   snowflake.ID
	RefCount int
}

type FileToken struct {
	ID          snowflake.ID
	OwnershipID snowflake.ID
	FileID      string
	UserID      snowflake.ID
	Bucket      string
	Size        int
	Type        string
	ExpiresAt   time.Time
}

type FileDomain struct {
	snowflake            *snowflake.Node
	fileTokenExpiration  time.Duration
	fileUploadExpiration time.Duration

	imageBucketName string
	otherBucketName string
}

func NewFileDomain(
	snowflake *snowflake.Node,
	fileTokenExpiration time.Duration,
	fileUploadExpiration time.Duration,
	imageBucketName string,
	otherBucketName string,
) *FileDomain {
	return &FileDomain{
		snowflake:            snowflake,
		fileTokenExpiration:  fileTokenExpiration,
		fileUploadExpiration: fileUploadExpiration,
		imageBucketName:      imageBucketName,
		otherBucketName:      otherBucketName,
	}
}

func (domain *FileDomain) NewUploadPolicy(userID snowflake.ID, allowedTypes []string, maxSize int64) *UploadPolicy {
	return &UploadPolicy{
		Token:        xcrypto.RandToken(),
		UserID:       userID,
		AllowedTypes: allowedTypes,
		MaxSize:      maxSize,
		ExpiresAt:    time.Now().Add(domain.fileUploadExpiration),
	}
}

func (domain *FileDomain) ClassifyBucket(t string) string {
	if mime.IsImage(t) {
		return domain.imageBucketName
	}

	return domain.otherBucketName
}

func (domain *FileDomain) NewFileInfo(id string, metadata *FileMetadata) *FileInfo {
	return &FileInfo{
		ID:        id,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}
}

func (domain *FileDomain) NewFileOwnership(fileID string, userID snowflake.ID) *FileOwnership {
	return &FileOwnership{
		ID:       domain.snowflake.Generate(),
		FileID:   fileID,
		UserID:   userID,
		RefCount: 0,
	}
}

func (domain *FileDomain) NewFileToken(info *FileInfo, ownership *FileOwnership) *FileToken {
	return &FileToken{
		ID:          domain.snowflake.Generate(),
		OwnershipID: ownership.ID,
		FileID:      ownership.FileID,
		UserID:      ownership.UserID,
		Bucket:      info.Metadata.Bucket,
		Size:        info.Metadata.Size,
		Type:        info.Metadata.Type,
		ExpiresAt:   time.Now().Add(domain.fileTokenExpiration),
	}
}
