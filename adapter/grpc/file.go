package grpc

import (
	"context"

	"github.com/todennus/file-service/adapter/abstraction"
	"github.com/todennus/file-service/adapter/grpc/conversion"
	"github.com/todennus/proto/gen/service"
	pbdto "github.com/todennus/proto/gen/service/dto"
	"github.com/todennus/shared/errordef"
	"github.com/todennus/shared/interceptor"
	"github.com/todennus/shared/response"
	"google.golang.org/grpc/codes"
)

type FileServer struct {
	fileUsecase abstraction.FileUsecase
	service.UnimplementedFileServer
}

func NewFileServer(fileUsecase abstraction.FileUsecase) *FileServer {
	return &FileServer{fileUsecase: fileUsecase}
}

func (server *FileServer) RegisterUpload(
	ctx context.Context,
	req *pbdto.FileRegisterUploadRequest,
) (*pbdto.FileRegisterUploadResponse, error) {
	if err := interceptor.RequireAuthentication(ctx); err != nil {
		return nil, err
	}

	resp, err := server.fileUsecase.RegisterUpload(ctx, conversion.NewUsecaseRegisterUploadRequest(req))
	return response.NewGRPCResponseHandler(ctx, conversion.NewPbFileRegisterUploadResponse(resp), err).
		Map(codes.PermissionDenied, errordef.ErrForbidden).
		Map(codes.InvalidArgument, errordef.ErrRequestInvalid).
		Finalize(ctx)
}

func (server *FileServer) CreatePresignedURL(
	ctx context.Context,
	req *pbdto.FileCreatePresignedURLRequest,
) (*pbdto.FileCreatePresignedURLResponse, error) {
	if err := interceptor.RequireAuthentication(ctx); err != nil {
		return nil, err
	}

	resp, err := server.fileUsecase.CreatePresignedURL(ctx, conversion.NewUsecaseCreatePresignedURLRequest(req))
	return response.NewGRPCResponseHandler(ctx, conversion.NewPbFileCreatePresignedURLResponse(resp), err).
		Map(codes.InvalidArgument, errordef.ErrRequestInvalid).
		Map(codes.PermissionDenied, errordef.ErrForbidden).
		Finalize(ctx)
}

func (server *FileServer) ChangeRefcount(
	ctx context.Context, req *pbdto.FileChangeRefcountRequest,
) (*pbdto.FileChangeRefcountResponse, error) {
	if err := interceptor.RequireAuthentication(ctx); err != nil {
		return nil, err
	}

	resp, err := server.fileUsecase.ChangeRefCount(ctx, conversion.NewUsecaseChangeRefcountRequest(req))
	return response.NewGRPCResponseHandler(ctx, conversion.NewPbChangeRefcountResponse(resp), err).
		Map(codes.InvalidArgument, errordef.ErrRequestInvalid).
		Map(codes.PermissionDenied, errordef.ErrForbidden).
		Finalize(ctx)
}
