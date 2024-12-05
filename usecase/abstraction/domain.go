package abstraction

import (
	"github.com/todennus/file-service/domain"
	"github.com/xybor-x/snowflake"
)

type FileDomain interface {
	ClassifyBucket(t string) string
	NewUploadPolicy(userID snowflake.ID, allowedTypes []string, maxSize int64) *domain.UploadPolicy
	NewFileInfo(id string, metadata *domain.FileMetadata) *domain.FileInfo
	NewFileOwnership(fileID string, userID snowflake.ID) *domain.FileOwnership
	NewFileToken(file *domain.FileInfo, ownership *domain.FileOwnership) *domain.FileToken
}
