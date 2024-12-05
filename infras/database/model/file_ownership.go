package model

import (
	"github.com/todennus/file-service/domain"
	"github.com/xybor-x/snowflake"
)

type FileOwnership struct {
	ID       int64  `gorm:"column:id;primaryKey"`
	FileID   string `gorm:"column:file_id"`
	UserID   int64  `gorm:"column:user_id"`
	RefCount int    `gorm:"column:refcount"`
}

func (FileOwnership) TableName() string {
	return "file_ownerships"
}

func NewFileOwnership(f *domain.FileOwnership) *FileOwnership {
	return &FileOwnership{
		ID:       f.ID.Int64(),
		FileID:   f.FileID,
		UserID:   f.UserID.Int64(),
		RefCount: f.RefCount,
	}
}

func (f *FileOwnership) To() *domain.FileOwnership {
	return &domain.FileOwnership{
		ID:       snowflake.ID(f.ID),
		FileID:   f.FileID,
		UserID:   snowflake.ID(f.UserID),
		RefCount: f.RefCount,
	}
}
