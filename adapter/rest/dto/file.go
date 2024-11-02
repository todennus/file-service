package dto

import (
	"github.com/todennus/file-service/usecase/dto"
	"github.com/todennus/shared/errordef"
	"github.com/todennus/x/xerror"
	"github.com/todennus/x/xhttp"
)

type ValidatePolicyRequest struct {
	PolicyToken string `json:"policy_token"`
	Type        string `json:"type"`
	Size        int    `json:"size"`
}

func (req *ValidatePolicyRequest) To() *dto.ValidatePolicyRequest {
	return &dto.ValidatePolicyRequest{
		PolicyToken: req.PolicyToken,
		Type:        req.Type,
		Size:        req.Size,
	}
}

type ValidatePolicyResponse struct {
	UploadToken string `json:"upload_token"`
}

func NewValidatePolicyResponse(resp *dto.ValidatePolicyResponse) *ValidatePolicyResponse {
	if resp == nil {
		return nil
	}

	return &ValidatePolicyResponse{
		UploadToken: resp.UploadToken,
	}
}

var _ xhttp.FormDataRequest = (*UploadRequest)(nil)

type UploadRequest struct {
	UploadToken string                `multipart:"upload_token"`
	File        xhttp.SniffReadCloser `multipart:"file,filesniff"`
	Size        int                   `multipart:"file,filesize"`
}

func (req *UploadRequest) To() (*dto.UploadRequest, error) {
	if req.File == nil {
		return nil, xerror.Enrich(errordef.ErrRequestInvalid, "not found file")
	}

	return &dto.UploadRequest{
		UploadToken: req.UploadToken,
		Content:     req.File,
		Size:        req.Size,
	}, nil
}

func (req *UploadRequest) NumFiles() int {
	return 1
}

type UploadResponse struct {
	TemporaryFileToken string `json:"temporary_file_token"`
}

func NewUploadResponse(resp *dto.UploadResponse) *UploadResponse {
	if resp == nil {
		return nil
	}

	return &UploadResponse{
		TemporaryFileToken: resp.TemporaryFileToken,
	}
}
