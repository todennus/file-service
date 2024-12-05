package abstraction

import (
	"context"
	"io"
	"time"

	"github.com/todennus/file-service/domain"
	"github.com/xybor-x/snowflake"
)

type FileUploadPolicyRepository interface {
	Save(ctx context.Context, policy *domain.UploadPolicy) error
	LoadAndDelete(ctx context.Context, token string) (*domain.UploadPolicy, error)
}

type FileInfoRepository interface {
	Create(ctx context.Context, file *domain.FileInfo) error
	GetByID(ctx context.Context, id string) (*domain.FileInfo, error)
}

type FileOwnershipRepository interface {
	Create(ctx context.Context, fileowner *domain.FileOwnership) error
	Get(ctx context.Context, fileID string, userID snowflake.ID) (*domain.FileOwnership, error)
	GetByID(ctx context.Context, id snowflake.ID) (*domain.FileOwnership, error)
	ChangeRefCount(ctx context.Context, id snowflake.ID, change int) error
}

type FileStorageRepository interface {
	Presign(ctx context.Context, file *domain.FileInfo, expiration time.Duration) (string, error)
	Store(ctx context.Context, file *domain.FileInfo, content io.Reader) error
}
