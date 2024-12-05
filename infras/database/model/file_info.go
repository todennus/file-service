package model

import (
	"time"

	"github.com/todennus/file-service/domain"
)

type FileInfo struct {
	ID        string `gorm:"column:id;primaryKey"`
	Bucket    string `gorm:"column:bucket"`
	Type      string `gorm:"column:type"`
	Size      int    `gorm:"column:size"`
	CreatedAt time.Time
}

func (FileInfo) TableName() string {
	return "files"
}

func NewFileInfo(f *domain.FileInfo) *FileInfo {
	return &FileInfo{
		ID:        f.ID,
		Bucket:    f.Metadata.Bucket,
		Type:      f.Metadata.Type,
		Size:      f.Metadata.Size,
		CreatedAt: f.CreatedAt,
	}
}

func (f *FileInfo) To() *domain.FileInfo {
	return &domain.FileInfo{
		ID: f.ID,
		Metadata: &domain.FileMetadata{
			Bucket: f.Bucket,
			Type:   f.Type,
			Size:   f.Size,
		},
		CreatedAt: f.CreatedAt,
	}
}
