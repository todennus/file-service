package abstraction

import (
	"context"
	"io"

	"github.com/todennus/file-service/domain"
	"github.com/todennus/file-service/usecase/dto"
)

type UserRepository interface {
	ValidateAvatarPolicyToken(ctx context.Context, policyToken string) (*dto.OverridenPolicyInfo, error)
}

type FileSessionRepository interface {
	SaveUploadSession(ctx context.Context, session *domain.UploadSession) error
	LoadUploadSession(ctx context.Context, uploadToken string) (*domain.UploadSession, error)
	DeleteUploadSession(ctx context.Context, uploadToken string) error

	SaveTemporarySession(ctx context.Context, session *domain.TemporaryFileSession) error
	LoadTemporarySession(ctx context.Context, sessionToken string) (*domain.TemporaryFileSession, error)
	DeleteTemporarySession(ctx context.Context, sessionToken string) error
}

type FileStorageRepository interface {
	ImageURL(name string) string
	ImageExists(ctx context.Context, name string) (bool, error)
	ImageSave(ctx context.Context, name string, temporaryFileName string) error

	StoreTemporary(
		ctx context.Context,
		temporaryFileName string,
		content io.Reader,
		size int64,
		contentType string,
	) (*dto.StorageUploadFileMetadata, error)
	GetTemporary(ctx context.Context, temporaryFileName string) (io.ReadCloser, *dto.StorageDownloadFileMetadata, error)
	RemoveTemporary(ctx context.Context, temporaryFileName string) error
}
