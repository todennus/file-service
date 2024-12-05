package dto

import (
	"github.com/todennus/file-service/usecase/dto"
	"github.com/todennus/shared/errordef"
	"github.com/todennus/x/xerror"
	"github.com/todennus/x/xhttp"
	"github.com/xybor-x/snowflake"
)

var _ xhttp.FormDataRequest = (*UploadRequest)(nil)

type UploadRequest struct {
	UploadToken string      `multipart:"upload_token"`
	File        *xhttp.File `multipart:"file,file"`
}

func (req *UploadRequest) To() (*dto.UploadRequest, error) {
	if req.File == nil {
		return nil, xerror.Enrich(errordef.ErrRequestInvalid, "not found file")
	}

	return &dto.UploadRequest{
		UploadToken: req.UploadToken,
		File:        req.File,
	}, nil
}

func (req *UploadRequest) NumFiles() int {
	return 1
}

type UploadResponse struct {
	Bucket      string `json:"bucket"`
	FileID      string `json:"file_id"`
	OwnershipID string `json:"ownership_id"`
	FileToken   string `json:"file_token"`
}

func NewUploadResponse(resp *dto.UploadResponse) *UploadResponse {
	if resp == nil {
		return nil
	}

	return &UploadResponse{
		Bucket:      resp.Bucket,
		FileID:      resp.FileID,
		OwnershipID: resp.OwnershipID.String(),
		FileToken:   resp.FileToken,
	}
}

type RetrieveFileTokenRequest struct {
	OwnershipID int64 `param:"ownership_id"`
}

func (req *RetrieveFileTokenRequest) To() *dto.RetrieveFileTokenRequest {
	return &dto.RetrieveFileTokenRequest{OwnershipID: snowflake.ID(req.OwnershipID)}
}

type RetrieveFileTokenResponse struct {
	FileToken string `json:"file_token"`
}

func NewRetrieveFileTokenResponse(resp *dto.RetrieveFileTokenResponse) *RetrieveFileTokenResponse {
	if resp == nil {
		return nil
	}

	return &RetrieveFileTokenResponse{
		FileToken: resp.FileToken,
	}
}
