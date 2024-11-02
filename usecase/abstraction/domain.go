package abstraction

import "github.com/todennus/file-service/domain"

type FileDomain interface {
	DefaultAvatarUploadPolicy() *domain.UploadPolicy
	NewUploadSession(source string, sourceInfo string, ftype string, fsize int) *domain.UploadSession
	NewTemporaryFileSession(info *domain.UploadSession) *domain.TemporaryFileSession
}
