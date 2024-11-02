package model

import (
	"time"

	"github.com/todennus/file-service/domain"
)

type FileMetadata struct {
	Type string `json:"typ"`
	Size int    `json:"sze"`
}

func NewFileMetadata(m *domain.FileMetadata) *FileMetadata {
	return &FileMetadata{
		Type: m.Type,
		Size: m.Size,
	}
}

func (m *FileMetadata) To() *domain.FileMetadata {
	return &domain.FileMetadata{
		Type: m.Type,
		Size: m.Size,
	}
}

type UploadSessionRecord struct {
	Source     string `json:"src"`
	SourceInfo string `json:"sif,omitempty"`
	ExpiresAt  int64  `json:"exp,omitempty"`
	*FileMetadata
}

func NewUploadSessionRecord(session *domain.UploadSession) *UploadSessionRecord {
	return &UploadSessionRecord{
		Source:       session.PolicySource,
		SourceInfo:   session.PolicyMetadata,
		FileMetadata: NewFileMetadata(session.FileMetadata),
		ExpiresAt:    session.ExpiresAt.Unix(),
	}
}

func (record *UploadSessionRecord) To(uploadToken string) *domain.UploadSession {
	return &domain.UploadSession{
		Token:          uploadToken,
		PolicySource:   record.Source,
		PolicyMetadata: record.SourceInfo,
		FileMetadata:   record.FileMetadata.To(),
		ExpiresAt:      time.Unix(record.ExpiresAt, 0),
	}
}

type TemporaryFileSessionRecord struct {
	UploadSessionInfo *UploadSessionRecord `json:"inf"`
	FileHash          string               `json:"fhs"`
	ExpiresAt         int64                `json:"exp"`
}

func NewTemporaryFileSessionRecord(session *domain.TemporaryFileSession) *TemporaryFileSessionRecord {
	return &TemporaryFileSessionRecord{
		UploadSessionInfo: NewUploadSessionRecord(session.UploadSessionInfo),
		FileHash:          session.FileHash,
		ExpiresAt:         session.ExpiresAt.Unix(),
	}
}

func (record *TemporaryFileSessionRecord) To(temporaryFileToken string) *domain.TemporaryFileSession {
	return &domain.TemporaryFileSession{
		Token:             temporaryFileToken,
		UploadSessionInfo: record.UploadSessionInfo.To(""),
		FileHash:          record.FileHash,
		ExpiresAt:         time.Unix(record.ExpiresAt, 0),
	}
}
