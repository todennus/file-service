package wiring

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"github.com/todennus/migration/postgres"
	"github.com/todennus/shared/config"
	"gorm.io/gorm"
)

type Infras struct {
	GormPostgres *gorm.DB
	Redis        *redis.Client
	Minio        *minio.Client
}

func InitializeInfras(ctx context.Context, config *config.Config) (*Infras, error) {
	infras := Infras{}
	var err error

	infras.GormPostgres, err = postgres.Initialize(ctx, config)
	if err != nil {
		return nil, err
	}

	infras.Redis = redis.NewClient(&redis.Options{
		Addr:     config.Variable.Redis.Addr,
		DB:       config.Variable.Redis.DB,
		Username: config.Secret.Redis.Username,
		Password: config.Secret.Redis.Password,
	})

	minioOpts := minio.Options{
		Creds: credentials.NewStaticV4(config.Secret.Minio.AccessKey, config.Secret.Minio.SecretKey, ""),
	}

	if infras.Minio, err = minio.New(config.Variable.Minio.Endpoint, &minioOpts); err != nil {
		return nil, err
	}

	return &infras, nil
}
