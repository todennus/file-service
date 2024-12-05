package abstraction

import (
	"context"

	"github.com/todennus/file-service/usecase/dto"
)

type FileUsecase interface {
	RegisterUpload(context.Context, *dto.RegisterUploadRequest) (*dto.RegisterUploadResponse, error)
	CreatePresignedURL(context.Context, *dto.CreatePresignedURLRequest) (*dto.CreatePresignedURLResponse, error)
	ChangeRefCount(context.Context, *dto.ChangeRefcountRequest) (*dto.ChangeRefcountResponse, error)

	Upload(context.Context, *dto.UploadRequest) (*dto.UploadResponse, error)
	RetrieveFileToken(context.Context, *dto.RetrieveFileTokenRequest) (*dto.RetrieveFileTokenResponse, error)
}
