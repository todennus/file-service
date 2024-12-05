package postgres

import (
	"context"

	"github.com/todennus/file-service/domain"
	"github.com/todennus/file-service/infras/database/model"
	"github.com/todennus/shared/errordef"
	"github.com/todennus/shared/xcontext"
	"gorm.io/gorm"
)

type FileInfoRepository struct {
	db *gorm.DB
}

func NewFileInfoRepository(db *gorm.DB) *FileInfoRepository {
	return &FileInfoRepository{db: db}
}

func (repo *FileInfoRepository) Create(ctx context.Context, file *domain.FileInfo) error {
	return errordef.ConvertGormError(
		xcontext.DB(ctx, repo.db).Create(model.NewFileInfo(file)).Error,
	)
}

func (repo *FileInfoRepository) GetByID(ctx context.Context, fileID string) (*domain.FileInfo, error) {
	model := model.FileInfo{}
	if err := xcontext.DB(ctx, repo.db).Take(&model, "id=?", fileID).Error; err != nil {
		return nil, errordef.ConvertGormError(err)
	}

	return model.To(), nil
}
