package dto

import (
	"encoding/json"

	"github.com/todennus/shared/enumdef"
	"github.com/todennus/x/xhttp"
)

type ValidatePolicyRequest struct {
	PolicyToken string
	Type        string
	Size        int
}

type ValidatePolicyResponse struct {
	UploadToken string
}

func NewValidatePolicyResponse(updateToken string) *ValidatePolicyResponse {
	return &ValidatePolicyResponse{UploadToken: updateToken}
}

type UploadRequest struct {
	UploadToken string
	Content     xhttp.SniffReadCloser
	Size        int
}

type UploadResponse struct {
	TemporaryFileToken string
}

func NewUploadResponse(sessionToken string) *UploadResponse {
	return &UploadResponse{TemporaryFileToken: sessionToken}
}

type ValidateTemporaryFileRequest struct {
	TemporaryFileToken string
}

type ValidateTemporaryFileResponse struct {
	PolicyMetadata string
	Type           string
	Size           int
}

func NewValidateTemporaryFileResponse(
	policyMetadata string,
	ftype string,
	fsize int,
) *ValidateTemporaryFileResponse {
	return &ValidateTemporaryFileResponse{
		PolicyMetadata: policyMetadata,
		Type:           ftype,
		Size:           fsize,
	}
}

type CommandTemporaryFileRequest struct {
	TemporaryFileToken string
	Command            enumdef.TemporaryFileCommand
	Metadata           string
	PolicySource       string
}

type CommandTemporaryFileResponse struct {
	// Only available if the command is "save_*".
	PersistentURL string

	// Only available for commands which modify the file content.
	NextTemporaryFileToken string

	// Result is available for read command. It should be serialized as a string.
	Result string
}

func NewTemporaryFileCommandSave(persistentURL string) *CommandTemporaryFileResponse {
	return &CommandTemporaryFileResponse{PersistentURL: persistentURL}
}

func NewCommandTemporaryFileRead(data any) (*CommandTemporaryFileResponse, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &CommandTemporaryFileResponse{Result: string(b)}, nil
}
