package conversion

import (
	"time"

	ucdto "github.com/todennus/file-service/usecase/dto"
	pbdto "github.com/todennus/proto/gen/service/dto"
	"github.com/xybor-x/snowflake"
)

func NewUsecaseRegisterUploadRequest(req *pbdto.FileRegisterUploadRequest) *ucdto.RegisterUploadRequest {
	return &ucdto.RegisterUploadRequest{
		UserID:       snowflake.ParseInt64(req.GetUserId()),
		MaxSize:      req.GetMaxSize(),
		AllowedTypes: req.GetAllowedTypes(),
	}
}

func NewPbFileRegisterUploadResponse(resp *ucdto.RegisterUploadResponse) *pbdto.FileRegisterUploadResponse {
	if resp == nil {
		return nil
	}

	return &pbdto.FileRegisterUploadResponse{
		UploadToken: resp.UploadToken,
	}
}

func NewUsecaseCreatePresignedURLRequest(req *pbdto.FileCreatePresignedURLRequest) *ucdto.CreatePresignedURLRequest {
	return &ucdto.CreatePresignedURLRequest{
		FileID:      req.GetFileId(),
		OwnershipID: snowflake.ParseInt64(req.GetOwnershipId()),
		Expiration:  time.Duration(req.GetExpiration()) * time.Second,
	}
}

func NewPbFileCreatePresignedURLResponse(resp *ucdto.CreatePresignedURLResponse) *pbdto.FileCreatePresignedURLResponse {
	if resp == nil {
		return nil
	}

	return &pbdto.FileCreatePresignedURLResponse{
		PresignedUrl: resp.PresignedURL,
	}
}

func NewUsecaseChangeRefcountRequest(req *pbdto.FileChangeRefcountRequest) *ucdto.ChangeRefcountRequest {
	inc := make([]snowflake.ID, 0)
	for i := range req.IncOwnershipId {
		inc = append(inc, snowflake.ParseInt64(req.IncOwnershipId[i]))
	}

	dec := make([]snowflake.ID, 0)
	for i := range req.DecOwnershipId {
		dec = append(dec, snowflake.ParseInt64(req.DecOwnershipId[i]))
	}

	return &ucdto.ChangeRefcountRequest{
		IncOwnershipID: inc,
		DecOwnershipID: dec,
	}
}

func NewPbChangeRefcountResponse(resp *ucdto.ChangeRefcountResponse) *pbdto.FileChangeRefcountResponse {
	if resp == nil {
		return nil
	}

	return &pbdto.FileChangeRefcountResponse{}
}
