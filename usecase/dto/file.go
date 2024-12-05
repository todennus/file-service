package dto

import (
	"time"

	"github.com/todennus/x/xhttp"
	"github.com/xybor-x/snowflake"
)

type RegisterUploadRequest struct {
	UserID       snowflake.ID
	AllowedTypes []string
	MaxSize      int64
}

type RegisterUploadResponse struct {
	UploadToken string
}

func NewRegisterUploadResponse(uploadToken string) *RegisterUploadResponse {
	return &RegisterUploadResponse{UploadToken: uploadToken}
}

type UploadRequest struct {
	UploadToken string
	File        *xhttp.File
}

type UploadResponse struct {
	FileID      string
	Bucket      string
	OwnershipID snowflake.ID
	FileToken   string
}

func NewUploadResponse(fileID, bucket string, ownershipID snowflake.ID, fileToken string) *UploadResponse {
	return &UploadResponse{FileID: fileID, Bucket: bucket, OwnershipID: ownershipID, FileToken: fileToken}
}

type RetrieveFileTokenRequest struct {
	OwnershipID snowflake.ID
}

type RetrieveFileTokenResponse struct {
	FileToken string
}

func NewRetrieveFileTokenResponse(fileToken string) *RetrieveFileTokenResponse {
	return &RetrieveFileTokenResponse{
		FileToken: fileToken,
	}
}

type CreatePresignedURLRequest struct {
	OwnershipID snowflake.ID
	FileID      string
	Expiration  time.Duration
}

type CreatePresignedURLResponse struct {
	PresignedURL string
}

func NewCreatePresignedURLResponse(presignedURL string) *CreatePresignedURLResponse {
	return &CreatePresignedURLResponse{
		PresignedURL: presignedURL,
	}
}

type ChangeRefcountRequest struct {
	IncOwnershipID []snowflake.ID
	DecOwnershipID []snowflake.ID
}

type ChangeRefcountResponse struct{}

func NewChangeRefcountResponse() *ChangeRefcountResponse {
	return &ChangeRefcountResponse{}
}
