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
		config.Variable.File.DefaultImageAllowedTypes,
		config.Variable.File.DefaultMaxSize,
		time.Duration(config.Variable.File.UploadSessionExpiration)*time.Second,
		time.Duration(config.Variable.File.TemporaryFileExpiration)*time.Second,
	)

	return domains, nil
}
