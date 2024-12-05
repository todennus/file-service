package dto

import (
	"time"

	"github.com/todennus/file-service/domain"
	"github.com/todennus/shared/tokendef"
)

func FileTokenFromDomain(t *domain.FileToken) *tokendef.FileToken {
	return &tokendef.FileToken{
		ID:          t.ID.String(),
		OwnershipID: t.OwnershipID.String(),
		FileID:      t.FileID,
		UserID:      t.UserID.String(),
		Type:        t.Type,
		Size:        t.Size,
		ExpiresAt:   int(t.ExpiresAt.Unix()),
	}
}

func FileTokenToDomain(t *tokendef.FileToken) *domain.FileToken {
	return &domain.FileToken{
		ID:          t.SnowflakeID(),
		OwnershipID: t.SnowflakeOwnershipID(),
		FileID:      t.FileID,
		UserID:      t.SnowflakeUserID(),
		Type:        t.Type,
		Size:        t.Size,
		ExpiresAt:   time.Unix(int64(t.ExpiresAt), 0),
	}
}
