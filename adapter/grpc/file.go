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

func (server *FileServer) ValidateTemporaryFile(
	ctx context.Context,
	req *pbdto.FileValidateTemporaryFileRequest,
) (*pbdto.FileValidateTemporaryFileResponse, error) {
	if err := interceptor.RequireAuthentication(ctx); err != nil {
		return nil, err
	}

	resp, err := server.fileUsecase.ValidateTemporaryFile(
		ctx, conversion.NewUsecaseValidateTemporaryFileRequest(req))

	return response.NewGRPCResponseHandler(ctx, conversion.NewPbFileValidateTemporaryFileResponse(resp), err).
		Map(codes.PermissionDenied, errordef.ErrForbidden).
		Map(codes.InvalidArgument, errordef.ErrRequestInvalid).
		Finalize(ctx)
}

func (server *FileServer) CommandTemporaryFile(
	ctx context.Context,
	req *pbdto.FileCommandTemporaryFileRequest,
) (*pbdto.FileCommandTemporaryFileResponse, error) {
	if err := interceptor.RequireAuthentication(ctx); err != nil {
		return nil, err
	}

	resp, err := server.fileUsecase.CommandTemporaryFile(
		ctx, conversion.NewUsecaseCommandTemporaryFileRequest(req))

	return response.NewGRPCResponseHandler(ctx, conversion.NewPbFileCommandTemporaryFileResponse(resp), err).
		Map(codes.InvalidArgument, errordef.ErrRequestInvalid, errordef.ErrFileInvalidContent).
		Map(codes.PermissionDenied, errordef.ErrForbidden).
		Finalize(ctx)
}
