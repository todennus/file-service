package conversion

import (
	ucdto "github.com/todennus/file-service/usecase/dto"
	pbdto "github.com/todennus/proto/gen/service/dto"
	"github.com/todennus/shared/enumdef"
)

func NewUsecaseValidateTemporaryFileRequest(req *pbdto.FileValidateTemporaryFileRequest) *ucdto.ValidateTemporaryFileRequest {
	return &ucdto.ValidateTemporaryFileRequest{
		TemporaryFileToken: req.GetTemporaryFileToken(),
	}
}

func NewPbFileValidateTemporaryFileResponse(resp *ucdto.ValidateTemporaryFileResponse) *pbdto.FileValidateTemporaryFileResponse {
	if resp == nil {
		return nil
	}

	return &pbdto.FileValidateTemporaryFileResponse{
		PolicyMetadata: resp.PolicyMetadata,
		Type:           resp.Type,
		Size:           int32(resp.Size),
	}
}

func NewUsecaseCommandTemporaryFileRequest(req *pbdto.FileCommandTemporaryFileRequest) *ucdto.CommandTemporaryFileRequest {
	return &ucdto.CommandTemporaryFileRequest{
		Command:            enumdef.TemporaryFileCommandFromGRPC(req.GetCommand()),
		Metadata:           req.GetMetadata(),
		PolicySource:       req.GetPolicySource(),
		TemporaryFileToken: req.GetTemporaryFileToken(),
	}
}

func NewPbFileCommandTemporaryFileResponse(resp *ucdto.CommandTemporaryFileResponse) *pbdto.FileCommandTemporaryFileResponse {
	if resp == nil {
		return nil
	}

	return &pbdto.FileCommandTemporaryFileResponse{
		PersistentUrl:          resp.PersistentURL,
		NextTemporaryFileToken: resp.NextTemporaryFileToken,
		Result:                 resp.Result,
	}
}
