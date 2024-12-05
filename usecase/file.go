package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"slices"

	"github.com/todennus/file-service/domain"
	"github.com/todennus/file-service/usecase/abstraction"
	"github.com/todennus/file-service/usecase/dto"
	"github.com/todennus/shared/errordef"
	"github.com/todennus/shared/middleware"
	"github.com/todennus/shared/scopedef"
	"github.com/todennus/shared/xcontext"
	"github.com/todennus/x/mime"
	"github.com/todennus/x/token"
	"github.com/todennus/x/xcrypto"
	"github.com/todennus/x/xerror"
	"github.com/todennus/x/xhttp"
)

type FileUsecase struct {
	maxInMemory int64

	tokenEngine token.Engine

	fileDomain abstraction.FileDomain

	fileUploadPolicyRepo abstraction.FileUploadPolicyRepository
	fileInfoRepo         abstraction.FileInfoRepository
	fileOwnershipRepo    abstraction.FileOwnershipRepository
	fileStorageRepo      abstraction.FileStorageRepository
}

func NewFileUsecase(
	maxInMemory int64,
	tokenEngine token.Engine,
	fileDomain abstraction.FileDomain,
	fileUploadPolicyRepo abstraction.FileUploadPolicyRepository,
	fileRepo abstraction.FileInfoRepository,
	fileOwnerRepo abstraction.FileOwnershipRepository,
	fileStorageRepo abstraction.FileStorageRepository,
) *FileUsecase {
	return &FileUsecase{
		maxInMemory: maxInMemory,
		tokenEngine: tokenEngine,

		fileDomain: fileDomain,

		fileUploadPolicyRepo: fileUploadPolicyRepo,
		fileInfoRepo:         fileRepo,
		fileOwnershipRepo:    fileOwnerRepo,
		fileStorageRepo:      fileStorageRepo,
	}
}

func (usecase *FileUsecase) RegisterUpload(
	ctx context.Context,
	req *dto.RegisterUploadRequest,
) (*dto.RegisterUploadResponse, error) {
	if scopedef.Eval(xcontext.Scope(ctx)).RequireAdmin(scopedef.AdminRegisterFilePolicy).IsUnsatisfied() {
		return nil, xerror.Enrich(errordef.ErrForbidden, "insufficient scope")
	}

	policy := usecase.fileDomain.NewUploadPolicy(req.UserID, req.AllowedTypes, req.MaxSize)
	if err := usecase.fileUploadPolicyRepo.Save(ctx, policy); err != nil {
		return nil, errordef.ErrServer.Hide(err, "failed-to-save-upload-policy")
	}

	return dto.NewRegisterUploadResponse(policy.Token), nil
}

func (usecase *FileUsecase) Upload(ctx context.Context, req *dto.UploadRequest) (*dto.UploadResponse, error) {
	defer req.File.Close()

	if xcontext.RequestSubjectID(ctx) == 0 {
		return nil, xerror.Enrich(errordef.ErrUnauthenticated, middleware.RequireAuthenticationMessage)
	}

	file, metadata, err := usecase.checkAndParseFile(ctx, req.File, req.UploadToken)
	if err != nil {
		return nil, err
	}

	fileHash, err := xcrypto.Sha256(file)
	if err != nil {
		return nil, errordef.ErrServer.Hide(err, "failed-to-hash-file")
	}

	fileInfo := usecase.fileDomain.NewFileInfo(base64.RawURLEncoding.EncodeToString(fileHash), metadata)

	ctx = xcontext.WithDBTransaction(ctx)
	if err := usecase.fileInfoRepo.Create(ctx, fileInfo); err == nil {
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			ctx = xcontext.DBRollback(ctx)
			return nil, errordef.ErrServer.Hide(err, "failed-to-seek-file")
		}

		if err := usecase.fileStorageRepo.Store(ctx, fileInfo, file); err != nil {
			ctx = xcontext.DBRollback(ctx)
			return nil, errordef.ErrServer.Hide(err, "failed-to-store-file")
		}
	} else if !errors.Is(err, errordef.ErrDuplicated) {
		ctx = xcontext.DBRollback(ctx)
		return nil, errordef.ErrServer.Hide(err, "failed-to-create-file-info")
	}

	// The transaction of storing the file must be committed before storing the
	// file's ownership.
	//
	// Why? It prevents the situation where the file is uploaded again, wasting
	// the server's resources.
	//
	// Don't worry if no one uses this file; it will be deleted periodically
	// by the janitor.
	ctx = xcontext.DBCommit(ctx)

	ownership := usecase.fileDomain.NewFileOwnership(fileInfo.ID, xcontext.RequestSubjectID(ctx))
	if err := usecase.fileOwnershipRepo.Create(ctx, ownership); err != nil && !errors.Is(err, errordef.ErrDuplicated) {
		return nil, errordef.ErrServer.Hide(err, "failed-to-create-file-owner-info")
	}

	fileToken := usecase.fileDomain.NewFileToken(fileInfo, ownership)
	fileTokenString, err := usecase.tokenEngine.Generate(ctx, dto.FileTokenFromDomain(fileToken))
	if err != nil {
		return nil, errordef.ErrServer.Hide(err, "failed-to-generate-file-token")
	}

	return dto.NewUploadResponse(fileInfo.ID, fileInfo.Metadata.Bucket, ownership.ID, fileTokenString), nil
}

func (usecase *FileUsecase) RetrieveFileToken(
	ctx context.Context,
	req *dto.RetrieveFileTokenRequest,
) (*dto.RetrieveFileTokenResponse, error) {
	ownership, err := usecase.fileOwnershipRepo.GetByID(ctx, req.OwnershipID)
	if err != nil {
		if errors.Is(err, errordef.ErrNotFound) {
			return nil, xerror.Enrich(errordef.ErrForbidden, "not found file ownership")
		}

		return nil, errordef.ErrServer.Hide(err, "failed-to-get-file-ownership")
	}

	if ownership.UserID != xcontext.RequestSubjectID(ctx) {
		return nil, xerror.Enrich(errordef.ErrForbidden, "the file is not owned by this user")
	}

	file, err := usecase.fileInfoRepo.GetByID(ctx, ownership.FileID)
	if err != nil {
		return nil, errordef.ErrServer.Hide(err, "failed-to-get-file-info")
	}

	fileToken := usecase.fileDomain.NewFileToken(file, ownership)
	fileTokenString, err := usecase.tokenEngine.Generate(ctx, dto.FileTokenFromDomain(fileToken))
	if err != nil {
		return nil, errordef.ErrServer.Hide(err, "failed-to-generate-file-token")
	}

	return dto.NewRetrieveFileTokenResponse(fileTokenString), nil
}

func (usecase *FileUsecase) CreatePresignedURL(
	ctx context.Context,
	req *dto.CreatePresignedURLRequest,
) (*dto.CreatePresignedURLResponse, error) {
	if scopedef.Eval(xcontext.Scope(ctx)).RequireAdmin(scopedef.AdminCreatePresignedFile).IsUnsatisfied() {
		return nil, xerror.Enrich(errordef.ErrForbidden, "insufficient scope")
	}

	if req.FileID != "" && req.OwnershipID != 0 {
		return nil, xerror.Enrich(errordef.ErrRequestInvalid, "not allow providing both file_id and ownership_id")
	}

	if req.Expiration == 0 {
		return nil, xerror.Enrich(errordef.ErrRequestInvalid, "require expiration")
	}

	fileID := req.FileID
	if fileID == "" {
		ownership, err := usecase.fileOwnershipRepo.GetByID(ctx, req.OwnershipID)
		if err != nil {
			return nil, errordef.ErrServer.Hide(err, "failed-to-get-file-ownership", "id", req.OwnershipID)
		}

		fileID = ownership.FileID
	}

	info, err := usecase.fileInfoRepo.GetByID(ctx, fileID)
	if err != nil {
		if errors.Is(err, errordef.ErrNotFound) {
			return nil, xerror.Enrich(errordef.ErrNotFound, "not found file %s", fileID)
		}

		return nil, errordef.ErrServer.Hide(err, "failed-to-get-file", "id", fileID)
	}

	presignedURL, err := usecase.fileStorageRepo.Presign(ctx, info, req.Expiration)
	if err != nil {
		return nil, errordef.ErrServer.Hide(err, "failed-to-generate-presigned-url")
	}

	return dto.NewCreatePresignedURLResponse(presignedURL), nil
}

func (usecase *FileUsecase) ChangeRefCount(
	ctx context.Context,
	req *dto.ChangeRefcountRequest,
) (*dto.ChangeRefcountResponse, error) {
	if scopedef.Eval(xcontext.Scope(ctx)).RequireAdmin(scopedef.AdminChangeRefcountFileOwnership).IsUnsatisfied() {
		return nil, xerror.Enrich(errordef.ErrForbidden, "insufficient scope")
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.DBCommit(ctx)

	for i := range req.IncOwnershipID {
		if err := usecase.fileOwnershipRepo.ChangeRefCount(ctx, req.IncOwnershipID[i], 1); err != nil {
			ctx = xcontext.DBRollback(ctx)
			return nil, errordef.ErrServer.Hide(err, "failed-to-increase-ref-count", "oid", req.IncOwnershipID[i])
		}
	}

	for i := range req.DecOwnershipID {
		if err := usecase.fileOwnershipRepo.ChangeRefCount(ctx, req.DecOwnershipID[i], -1); err != nil {
			ctx = xcontext.DBRollback(ctx)
			return nil, errordef.ErrServer.Hide(err, "failed-to-decrease-ref-count", "oid", req.DecOwnershipID[i])
		}
	}

	return dto.NewChangeRefcountResponse(), nil
}

func (usecase *FileUsecase) checkAndParseFile(
	ctx context.Context,
	file *xhttp.File,
	uploadToken string,
) (io.ReadSeeker, *domain.FileMetadata, error) {
	policy, err := usecase.fileUploadPolicyRepo.LoadAndDelete(ctx, uploadToken)
	if err != nil {
		if errors.Is(err, errordef.ErrNotFound) {
			return nil, nil, xerror.Enrich(errordef.ErrRequestInvalid, "invalid token")
		}

		return nil, nil, err
	}

	if xcontext.RequestSubjectID(ctx) != policy.UserID {
		return nil, nil, xerror.Enrich(errordef.ErrForbidden, "not allow the user to use this token")
	}

	contentType, err := usecase.checkContentType(file, policy.AllowedTypes)
	if err != nil {
		return nil, nil, err
	}

	var fileContent io.ReadSeeker
	file.SetMaxSize(policy.MaxSize)
	if policy.MaxSize > usecase.maxInMemory {
		fileContent, err = file.AsFile()
	} else {
		fileContent, err = file.AsBytes()
	}

	if err != nil {
		if mberr := xhttp.AsMaxBytesError(err); mberr != nil {
			return nil, nil, xerror.Enrich(errordef.ErrRequestTooLarge, "file too large (limit %d)", policy.MaxSize)
		}

		return nil, nil, errordef.ErrServer.Hide(err, "failed-to-parse-file")
	}

	return fileContent, &domain.FileMetadata{
		Bucket: usecase.fileDomain.ClassifyBucket(contentType),
		Type:   contentType,
		Size:   file.ContentLength(),
	}, nil
}

func (usecase *FileUsecase) checkContentType(file *xhttp.File, allowedTypes []string) (string, error) {
	nSniff := int64(512)
	if mime.IsImage(allowedTypes...) {
		// Although http.DetectContentType considers at most 512 bytes to detect
		// the file type, detecting image types needs only first 12 bytes.
		nSniff = 12
	}

	detectedType, err := file.ContentType(nSniff)
	if err != nil {
		return detectedType, errordef.ErrServer.Hide(err, "failed-to-detect-content-type")
	}

	if !slices.Contains(allowedTypes, detectedType) {
		return detectedType, xerror.Enrich(errordef.ErrFileMismatchedType,
			"mismachted uploaded file type (got %s, expected %s)", detectedType, allowedTypes)
	}

	return detectedType, nil
}
