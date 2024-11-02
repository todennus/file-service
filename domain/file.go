package domain

import (
	"time"

	"github.com/todennus/x/xcrypto"
)

type UploadPolicy struct {
	// AllowedTypes specifies the permitted content types for the uploaded file.
	// It defines which MIME types are acceptable for file uploads.
	AllowedTypes []string

	// Maxsize defines the maximum size, in bytes, of the uploaded file.
	MaxSize int
}

// FileMetadata contains additional information about a file.
type FileMetadata struct {
	// Type represents the MIME type of the file.
	Type string

	// Size is the size of the file content in bytes.
	Size int
}

// UploadSession is created after the file metadata has been checked against
// the upload policy. Before this session expires, the user can upload a file
// that matches the metadata specified above.
type UploadSession struct {
	// Token is a secret random value that a user can use to authenticate the
	// upload of a file.
	Token string

	// PolicySource represents the origin of the policy issuer.
	PolicySource string

	// PolicyMetadata contains additional information of the policy source. This
	// information should be serialized as a string.
	PolicyMetadata string

	// FileMetadata contains the file metadata which has been checked against
	// the policy. The file uploaded during this session must match this
	// metadata.
	FileMetadata *FileMetadata

	// ExpiresAt is the time after which this session expires.
	ExpiresAt time.Time
}

// TemporaryFileSession is created after a file is uploaded successfully. The
// file will be stored in a temporary storage. The policy source can use this
// session to perform some actions on that file. The temporary file will be
// deleted when the session expires or upon receiving a delete command from the
// policy source.
type TemporaryFileSession struct {
	// Token is a secret random value used by the policy source to authenticate
	// actions on the file. This value also serves as the filename in temporary
	// storage.
	Token string

	// UploadSessionInfo contains information about the upload session.
	UploadSessionInfo *UploadSession

	// FileHash is the hash value of the file content. This value serves as the
	// filename in persistent storage.
	FileHash string

	// ExpiresAt is the time after which this session expires.
	ExpiresAt time.Time
}

type FileDomain struct {
	defaultImageAllowedTypes []string
	defaultMaxFileSize       int
	uploadSessionExpiration  time.Duration
	fileSessionExpiration    time.Duration
}

func NewFileDomain(
	defaultImageAllowedTypes []string,
	defaultMaxFileSize int,
	uploadSessionExpiration time.Duration,
	fileSessionExpiration time.Duration,
) *FileDomain {
	return &FileDomain{
		defaultMaxFileSize:      defaultMaxFileSize,
		uploadSessionExpiration: uploadSessionExpiration,
		fileSessionExpiration:   fileSessionExpiration,
	}
}

func (domain *FileDomain) DefaultAvatarUploadPolicy() *UploadPolicy {
	return domain.defaultImageUploadPolicy()
}

func (domain *FileDomain) NewUploadSession(source, metadata string, ftype string, fsize int) *UploadSession {
	return &UploadSession{
		Token:          xcrypto.RandToken(),
		PolicySource:   source,
		PolicyMetadata: metadata,
		FileMetadata: &FileMetadata{
			Type: ftype,
			Size: fsize,
		},
		ExpiresAt: time.Now().Add(domain.uploadSessionExpiration),
	}
}

func (domain *FileDomain) NewTemporaryFileSession(policy *UploadSession) *TemporaryFileSession {
	return &TemporaryFileSession{
		Token:             xcrypto.RandToken(),
		UploadSessionInfo: policy,
		ExpiresAt:         time.Now().Add(domain.fileSessionExpiration),
	}
}

func (domain *FileDomain) defaultImageUploadPolicy() *UploadPolicy {
	return &UploadPolicy{
		AllowedTypes: domain.defaultImageAllowedTypes,
		MaxSize:      domain.defaultMaxFileSize,
	}
}
