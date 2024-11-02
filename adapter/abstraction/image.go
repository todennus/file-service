package abstraction

import (
	"context"

	"github.com/todennus/file-service/usecase/dto"
)

type FileUsecase interface {
	ValidatePolicy(context.Context, *dto.ValidatePolicyRequest) (*dto.ValidatePolicyResponse, error)
	Upload(context.Context, *dto.UploadRequest) (*dto.UploadResponse, error)
	ValidateTemporaryFile(context.Context, *dto.ValidateTemporaryFileRequest) (*dto.ValidateTemporaryFileResponse, error)
	CommandTemporaryFile(context.Context, *dto.CommandTemporaryFileRequest) (*dto.CommandTemporaryFileResponse, error)
}
