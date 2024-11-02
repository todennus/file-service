package grpc

import (
	"strconv"

	"github.com/todennus/file-service/domain"
	ucdto "github.com/todennus/file-service/usecase/dto"
	pbdto "github.com/todennus/proto/gen/service/dto"
)

func NewUserValidateAvatarResponse(resp *pbdto.UserValidateAvatarPolicyTokenResponse) *ucdto.OverridenPolicyInfo {
	return &ucdto.OverridenPolicyInfo{
		PolicySourceMetadata: strconv.FormatInt(resp.GetUserId(), 10),
		UploadPolicy: &domain.UploadPolicy{
			AllowedTypes: resp.GetAllowedTypes(),
			MaxSize:      int(resp.GetMaxSize()),
		},
	}
}
