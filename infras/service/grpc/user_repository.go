package grpc

import (
	"context"

	ucdto "github.com/todennus/file-service/usecase/dto"
	"github.com/todennus/proto/gen/service"
	pbdto "github.com/todennus/proto/gen/service/dto"
	"github.com/todennus/shared/authentication"
	"github.com/todennus/shared/errordef"
	"google.golang.org/grpc"
)

type UserRepository struct {
	auth   *authentication.GrpcAuthorization
	client service.UserClient
}

func NewUserRepository(grpcConn *grpc.ClientConn, auth *authentication.GrpcAuthorization) *UserRepository {
	return &UserRepository{
		client: service.NewUserClient(grpcConn),
		auth:   auth,
	}
}

func (repo *UserRepository) ValidateAvatarPolicyToken(ctx context.Context, policyToken string) (*ucdto.OverridenPolicyInfo, error) {
	req := &pbdto.UserValidateAvatarPolicyTokenRequest{PolicyToken: policyToken}
	resp, err := repo.client.ValidateAvatarPolicyToken(repo.auth.Context(ctx), req)
	if err != nil {
		return nil, errordef.ConvertGRPCError(err)
	}

	return NewUserValidateAvatarResponse(resp), nil
}
