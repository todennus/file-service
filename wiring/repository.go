package wiring

import (
	"context"

	"github.com/todennus/file-service/infras/database/postgres"
	"github.com/todennus/file-service/infras/database/redis"
	"github.com/todennus/file-service/infras/storage"
	"github.com/todennus/file-service/usecase/abstraction"
	"github.com/todennus/shared/config"
)

type Repositories struct {
	abstraction.FileUploadPolicyRepository
	abstraction.FileInfoRepository
	abstraction.FileOwnershipRepository
	abstraction.FileStorageRepository
}

func InitializeRepositories(ctx context.Context, config *config.Config, infras *Infras) (*Repositories, error) {
	r := &Repositories{}

	r.FileUploadPolicyRepository = redis.NewFilePolicyRepository(infras.Redis)
	r.FileInfoRepository = postgres.NewFileInfoRepository(infras.GormPostgres)
	r.FileOwnershipRepository = postgres.NewFileOwnershipRepository(infras.GormPostgres)
	r.FileStorageRepository = storage.NewFileStorageRepository(infras.Minio)

	return r, nil
}
