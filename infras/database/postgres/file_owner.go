package postgres

import (
	"context"

	"github.com/todennus/file-service/domain"
	"github.com/todennus/file-service/infras/database/model"
	"github.com/todennus/shared/errordef"
	"github.com/todennus/shared/xcontext"
	"github.com/xybor-x/snowflake"
	"gorm.io/gorm"
)

type FileOwnershipRepository struct {
	db *gorm.DB
}

func NewFileOwnershipRepository(db *gorm.DB) *FileOwnershipRepository {
	return &FileOwnershipRepository{db: db}
}

func (repo *FileOwnershipRepository) Create(ctx context.Context, ownership *domain.FileOwnership) error {
	return errordef.ConvertGormError(
		xcontext.DB(ctx, repo.db).Create(model.NewFileOwnership(ownership)).Error,
	)
}

func (repo *FileOwnershipRepository) GetByID(ctx context.Context, ownershipID snowflake.ID) (*domain.FileOwnership, error) {
	model := model.FileOwnership{}
	if err := xcontext.DB(ctx, repo.db).Take(&model, "id=?", ownershipID).Error; err != nil {
		return nil, errordef.ConvertGormError(err)
	}

	return model.To(), nil
}

func (repo *FileOwnershipRepository) Get(ctx context.Context, fileID string, userID snowflake.ID) (*domain.FileOwnership, error) {
	model := model.FileOwnership{}
	if err := xcontext.DB(ctx, repo.db).Take(&model, "file_id=? AND user_id=?", fileID, userID).Error; err != nil {
		return nil, errordef.ConvertGormError(err)
	}

	return model.To(), nil
}

func (repo *FileOwnershipRepository) ChangeRefCount(ctx context.Context, ownershipID snowflake.ID, change int) error {
	return errordef.ConvertGormError(
		xcontext.DB(ctx, repo.db).
			Model(&model.FileOwnership{}).
			Where("id=?", ownershipID).
			Update("refcount", gorm.Expr("refcount+?", change)).Error,
	)
}
