package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/todennus/file-service/domain"
	"github.com/todennus/file-service/infras/database/model"
	"github.com/todennus/shared/errordef"
)

func filePolicyKey(token string) string {
	return fmt.Sprintf("file:upload_policy:%s", token)
}

type FilePolicyRepository struct {
	redis *redis.Client
}

func NewFilePolicyRepository(redis *redis.Client) *FilePolicyRepository {
	return &FilePolicyRepository{redis: redis}
}

func (repo *FilePolicyRepository) Save(ctx context.Context, policy *domain.UploadPolicy) error {
	record := model.NewUploadPolicy(policy)
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return errordef.ConvertRedisError(
		repo.redis.SetEx(
			ctx,
			filePolicyKey(policy.Token),
			recordJSON,
			time.Until(policy.ExpiresAt),
		).Err(),
	)
}

func (repo *FilePolicyRepository) LoadAndDelete(ctx context.Context, uploadToken string) (*domain.UploadPolicy, error) {
	recordJSON, err := repo.redis.GetDel(ctx, filePolicyKey(uploadToken)).Result()
	if err != nil {
		return nil, errordef.ConvertRedisError(err)
	}

	record := model.UploadPolicy{}
	if err := json.Unmarshal([]byte(recordJSON), &record); err != nil {
		return nil, err
	}

	return record.To(uploadToken), nil
}
