package wiring

import (
	"context"
	"time"

	"github.com/todennus/file-service/domain"
	"github.com/todennus/file-service/usecase/abstraction"
	"github.com/todennus/shared/config"
)

type Domains struct {
	abstraction.FileDomain
}

func InitializeDomains(ctx context.Context, config *config.Config) (*Domains, error) {
	domains := &Domains{}

	domains.FileDomain = domain.NewFileDomain(
		config.SnowflakeNode,
		time.Duration(config.Variable.File.TokenExpiration)*time.Second,
		time.Duration(config.Variable.File.UploadTokenExpiration)*time.Second,
		config.Variable.File.StorageImageBucket,
		config.Variable.File.StorageOtherBucket,
	)

	return domains, nil
}
