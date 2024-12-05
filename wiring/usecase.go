package wiring

import (
	"context"

	"github.com/todennus/file-service/adapter/abstraction"
	"github.com/todennus/file-service/usecase"
	"github.com/todennus/shared/config"
)

type Usecases struct {
	abstraction.FileUsecase
}

func InitializeUsecases(
	ctx context.Context,
	config *config.Config,
	infras *Infras,
	domains *Domains,
	repositories *Repositories,
) (*Usecases, error) {
	uc := &Usecases{}

	uc.FileUsecase = usecase.NewFileUsecase(
		config.Variable.File.MaxInMemory,
		config.TokenEngine,
		domains.FileDomain,
		repositories.FileUploadPolicyRepository,
		repositories.FileInfoRepository,
		repositories.FileOwnershipRepository,
		repositories.FileStorageRepository,
	)

	return uc, nil
}
