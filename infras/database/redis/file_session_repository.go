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

func uploadSessionKey(token string) string {
	return fmt.Sprintf("file:upload_session:%s", token)
}

func temporaryFileSessionKey(token string) string {
	return fmt.Sprintf("file:temporary_session:%s", token)
}

type FileSessionRepository struct {
	redis *redis.Client
}

func NewFileSessionRepository(redis *redis.Client) *FileSessionRepository {
	return &FileSessionRepository{redis: redis}
}

func (repo *FileSessionRepository) SaveUploadSession(
	ctx context.Context,
	uploadSession *domain.UploadSession,
) error {
	record := model.NewUploadSessionRecord(uploadSession)
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return errordef.ConvertRedisError(
		repo.redis.SetEx(
			ctx,
			uploadSessionKey(uploadSession.Token),
			recordJSON,
			time.Until(uploadSession.ExpiresAt),
		).Err(),
	)
}

func (repo *FileSessionRepository) LoadUploadSession(ctx context.Context, uploadToken string) (*domain.UploadSession, error) {
	recordJSON, err := repo.redis.Get(ctx, uploadSessionKey(uploadToken)).Result()
	if err != nil {
		return nil, errordef.ConvertRedisError(err)
	}

	record := &model.UploadSessionRecord{}
	if err := json.Unmarshal([]byte(recordJSON), record); err != nil {
		return nil, err
	}

	return record.To(uploadToken), nil
}

func (repo *FileSessionRepository) DeleteUploadSession(ctx context.Context, uploadToken string) error {
	return errordef.ConvertRedisError(repo.redis.Del(ctx, uploadSessionKey(uploadToken)).Err())
}

func (repo *FileSessionRepository) SaveTemporarySession(ctx context.Context, session *domain.TemporaryFileSession) error {
	record := model.NewTemporaryFileSessionRecord(session)
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return errordef.ConvertRedisError(
		repo.redis.SetEx(
			ctx,
			temporaryFileSessionKey(session.Token),
			recordJSON,
			time.Until(session.ExpiresAt),
		).Err(),
	)
}

func (repo *FileSessionRepository) LoadTemporarySession(ctx context.Context, sessionToken string) (*domain.TemporaryFileSession, error) {
	recordJSON, err := repo.redis.Get(ctx, temporaryFileSessionKey(sessionToken)).Result()
	if err != nil {
		return nil, errordef.ConvertRedisError(err)
	}

	record := &model.TemporaryFileSessionRecord{}
	if err := json.Unmarshal([]byte(recordJSON), record); err != nil {
		return nil, err
	}

	return record.To(sessionToken), nil
}

func (repo *FileSessionRepository) DeleteTemporarySession(ctx context.Context, sessionToken string) error {
	return errordef.ConvertRedisError(repo.redis.Del(ctx, temporaryFileSessionKey(sessionToken)).Err())
}
