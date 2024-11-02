package wiring

import (
	"context"

	"github.com/todennus/file-service/adapter/abstraction"
	"github.com/todennus/file-service/usecase"
)

type Usecases struct {
	abstraction.FileUsecase
}

func InitializeUsecases(
	ctx context.Context,
	infras *Infras,
	domains *Domains,
	repositories *Repositories,
) (*Usecases, error) {
	uc := &Usecases{}

	uc.FileUsecase = usecase.NewFileUsecase(
		domains.FileDomain,
		repositories.FileSessionRepository,
		repositories.FileStorageRepository,
		repositories.UserRepository,
	)

	return uc, nil
}
