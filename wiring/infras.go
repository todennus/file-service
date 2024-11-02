package wiring

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"github.com/todennus/migration/postgres"
	"github.com/todennus/shared/authentication"
	"github.com/todennus/shared/config"
	"github.com/todennus/shared/scopedef"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

type Infras struct {
	Auth         *authentication.GrpcAuthorization
	GormPostgres *gorm.DB
	Redis        *redis.Client
	Minio        *minio.Client
	UserGRPCConn *grpc.ClientConn
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

	infras.UserGRPCConn, err = grpc.NewClient(
		config.Variable.Service.UserGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	infras.Minio, err = minio.New(config.Variable.Minio.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(config.Secret.Minio.AccessKey, config.Secret.Minio.SecretKey, ""),
	})

	infras.Auth = authentication.NewGrpcAuthorization((&clientcredentials.Config{
		TokenURL:     config.Variable.Service.OAuth2TokenURL,
		ClientID:     config.Secret.Service.ClientID,
		ClientSecret: config.Secret.Service.ClientSecret,
		Scopes:       []string{scopedef.AdminValidateFilePolicy.Scope()},
	}).TokenSource)

	return &infras, nil
}
