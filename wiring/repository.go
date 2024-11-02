package wiring

import (
	"context"

	"github.com/todennus/file-service/infras/database/redis"
	"github.com/todennus/file-service/infras/service/grpc"
	"github.com/todennus/file-service/infras/storage"
	"github.com/todennus/file-service/usecase/abstraction"
	"github.com/todennus/shared/config"
)

type Repositories struct {
	abstraction.UserRepository
	abstraction.FileSessionRepository
	abstraction.FileStorageRepository
}

func InitializeRepositories(ctx context.Context, config *config.Config, infras *Infras) (*Repositories, error) {
	r := &Repositories{}

	r.UserRepository = grpc.NewUserRepository(infras.UserGRPCConn, infras.Auth)

	r.FileSessionRepository = redis.NewFileSessionRepository(infras.Redis)
	r.FileStorageRepository = storage.NewFileStorageRepository(
		infras.Minio,
		config.Variable.File.StorageImageBucket,
		config.Variable.File.StorageTemporaryBucket,
	)

	return r, nil
}
