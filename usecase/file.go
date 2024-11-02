package usecase

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/todennus/file-service/domain"
	"github.com/todennus/file-service/usecase/abstraction"
	"github.com/todennus/file-service/usecase/dto"
	"github.com/todennus/shared/enumdef"
	"github.com/todennus/shared/errordef"
	"github.com/todennus/shared/scopedef"
	"github.com/todennus/shared/xcontext"
	"github.com/todennus/x/mime"
	"github.com/todennus/x/xerror"
	"github.com/todennus/x/xhttp"
)

type FileUsecase struct {
	fileDomain abstraction.FileDomain

	fileSessionRepo abstraction.FileSessionRepository
	fileStorageRepo abstraction.FileStorageRepository
	userRepo        abstraction.UserRepository
}

func NewFileUsecase(
	fileDomain abstraction.FileDomain,
	fileSessionRepo abstraction.FileSessionRepository,
	fileStorageRepo abstraction.FileStorageRepository,
	userRepo abstraction.UserRepository,
) *FileUsecase {
	return &FileUsecase{
		fileDomain: fileDomain,

		fileSessionRepo: fileSessionRepo,
		fileStorageRepo: fileStorageRepo,
		userRepo:        userRepo,
	}
}

func (usecase *FileUsecase) ValidatePolicy(
	ctx context.Context,
	req *dto.ValidatePolicyRequest,
) (*dto.ValidatePolicyResponse, error) {
	if req.Size == 0 {
		return nil, xerror.Enrich(errordef.ErrRequestInvalid, "require a positive size")
	}

	source, _, found := enumdef.ParseFilePolicyToken(req.PolicyToken)
	if !found {
		return nil, xerror.Enrich(errordef.ErrRequestInvalid, "invalid policy token format")
	}

	// Get the overriden policy from the source.
	var err error
	var overridenPolicy *dto.OverridenPolicyInfo
	var policy *domain.UploadPolicy
	switch source {
	case enumdef.PolicySourceUserAvatar:
		policy = usecase.fileDomain.DefaultAvatarUploadPolicy()
		overridenPolicy, err = usecase.userRepo.ValidateAvatarPolicyToken(ctx, req.PolicyToken)
		if err != nil {
			return nil, usecase.translateValidateExternalPolicyError(err, source)
		}

	default:
		return nil, xerror.Enrich(errordef.ErrRequestInvalid, "invalid policy source")
	}

	// Override the default policy by the source policy, then validate the image
	// information against the policy.
	dto.OverridePolicyInfo(overridenPolicy, policy)
	if err := usecase.validatePolicy(req, policy); err != nil {
		return nil, err
	}

	// Generate an upload session. The user can use the upload token to upload
	// a file.
	session := usecase.fileDomain.NewUploadSession(
		source, overridenPolicy.PolicySourceMetadata, req.Type, req.Size)
	if err := usecase.fileSessionRepo.SaveUploadSession(ctx, session); err != nil {
		return nil, errordef.ErrServer.Hide(err, "failed-to-save-policy-session")
	}

	return dto.NewValidatePolicyResponse(session.Token), nil
}

func (usecase *FileUsecase) Upload(
	ctx context.Context,
	req *dto.UploadRequest,
) (*dto.UploadResponse, error) {
	defer req.Content.Close()

	// Load the policy uploadSession
	uploadSession, err := usecase.fileSessionRepo.LoadUploadSession(ctx, req.UploadToken)
	if err != nil {
		if errors.Is(err, errordef.ErrNotFound) {
			return nil, xerror.Enrich(errordef.ErrForbidden, "invalid upload token")
		}

		return nil, errordef.ErrServer.Hide(err, "failed-to-load-upload-session")
	}

	if err := usecase.fileSessionRepo.DeleteUploadSession(ctx, req.UploadToken); err != nil {
		xcontext.Logger(ctx).Warn("failed-to-delete-upload-session", "err", err)
	}

	if uploadSession.ExpiresAt.Before(time.Now()) {
		return nil, xerror.Enrich(errordef.ErrForbidden, "expired upload token")
	}

	// If the file size can be determined from the request, it should be checked
	// here. Otherwise, it will need to be checked after uploading the file to
	// storage.
	if req.Size != -1 && req.Size != uploadSession.FileMetadata.Size {
		return nil, xerror.Enrich(errordef.ErrFileMismatchedSize,
			"mismachted uploaded file size (got %d, expected %d)", req.Size, uploadSession.FileMetadata.Size)
	}

	nSniff := 512
	if mime.IsImage[uploadSession.FileMetadata.Type] {
		// Although http.DetectContentType considers at most 512 bytes to detect
		// the file type, detecting image types needs only first 12 bytes.
		nSniff = 12
	}

	detectedType := xhttp.DetectContentType(req.Content, nSniff)
	if detectedType != uploadSession.FileMetadata.Type {
		return nil, xerror.Enrich(errordef.ErrFileMismatchedType,
			"mismachted uploaded file type (got %s, expected %s)", detectedType, uploadSession.FileMetadata.Type)
	}

	temporaryFileSession := usecase.fileDomain.NewTemporaryFileSession(uploadSession)

	metadata, err := usecase.fileStorageRepo.StoreTemporary(
		ctx,
		temporaryFileSession.Token,
		http.MaxBytesReader(nil, req.Content, int64(uploadSession.FileMetadata.Size)),
		int64(req.Size),
		detectedType,
	)
	if err != nil {
		return nil, errordef.ErrServer.Hide(err, "failed-to-store-temporary-file")
	}

	shouldRemoveFile := false
	defer func() {
		if shouldRemoveFile {
			if err := usecase.fileStorageRepo.RemoveTemporary(ctx, temporaryFileSession.Token); err != nil {
				xcontext.Logger(ctx).Warn("failed-to-remove-invalid-temporary-file", "err", err)
			}
		}
	}()

	// If the file size cannot be determined from the request, it will be
	// checked after uploading the file to storage.
	if req.Size == -1 && metadata.Size != uploadSession.FileMetadata.Size {
		shouldRemoveFile = true
		return nil, xerror.Enrich(errordef.ErrRequestInvalid,
			"mismachted uploaded file size (got %d, expected %d)", metadata.Size, uploadSession.FileMetadata.Size)
	}

	temporaryFileSession.FileHash = metadata.Hash
	if err := usecase.fileSessionRepo.SaveTemporarySession(ctx, temporaryFileSession); err != nil {
		shouldRemoveFile = true
		return nil, errordef.ErrServer.Hide(err, "failed-to-save-temporary-file-session")
	}

	return dto.NewUploadResponse(temporaryFileSession.Token), nil
}

func (usecase *FileUsecase) ValidateTemporaryFile(
	ctx context.Context,
	req *dto.ValidateTemporaryFileRequest,
) (*dto.ValidateTemporaryFileResponse, error) {
	if scopedef.Eval(xcontext.Scope(ctx)).RequireAdmin(scopedef.AdminCommandTemporaryFile).IsUnsatisfied() {
		return nil, xerror.Enrich(errordef.ErrForbidden, "insufficient scope")
	}

	temporaryFileSession, err := usecase.fileSessionRepo.LoadTemporarySession(ctx, req.TemporaryFileToken)
	if err != nil {
		if errors.Is(err, errordef.ErrNotFound) {
			return nil, xerror.Enrich(errordef.ErrRequestInvalid, "invalid token")
		}

		return nil, errordef.ErrServer.Hide(err, "failed-to-load-temporary-file-session")
	}

	if temporaryFileSession.ExpiresAt.Before(time.Now()) {
		return nil, xerror.Enrich(errordef.ErrRequestInvalid, "expired token")
	}

	return dto.NewValidateTemporaryFileResponse(
		temporaryFileSession.UploadSessionInfo.PolicyMetadata,
		temporaryFileSession.UploadSessionInfo.FileMetadata.Type,
		temporaryFileSession.UploadSessionInfo.FileMetadata.Size,
	), nil
}

func (usecase *FileUsecase) CommandTemporaryFile(
	ctx context.Context,
	req *dto.CommandTemporaryFileRequest,
) (*dto.CommandTemporaryFileResponse, error) {
	if scopedef.Eval(xcontext.Scope(ctx)).RequireAdmin(scopedef.AdminCommandTemporaryFile).IsUnsatisfied() {
		return nil, xerror.Enrich(errordef.ErrForbidden, "insufficient scope")
	}

	session, err := usecase.fileSessionRepo.LoadTemporarySession(ctx, req.TemporaryFileToken)
	if err != nil {
		if errors.Is(err, errordef.ErrNotFound) {
			return nil, xerror.Enrich(errordef.ErrRequestInvalid, "invalid token")
		}

		return nil, errordef.ErrServer.Hide(err, "failed-to-load-temporary-file-session")
	}

	if err := usecase.fileSessionRepo.DeleteTemporarySession(ctx, req.TemporaryFileToken); err != nil {
		xcontext.Logger(ctx).Warn("failed-to-delete-temporary-file-session", "err", err)
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, xerror.Enrich(errordef.ErrRequestInvalid, "expired token")
	}

	if session.UploadSessionInfo.PolicySource != req.PolicySource {
		return nil, xerror.Enrich(errordef.ErrRequestInvalid, "mismatched policy source")
	}

	shouldRemoveTemporary := false
	var newContent *dto.FileContentWrapper
	var resp *dto.CommandTemporaryFileResponse
	switch req.Command {
	case enumdef.TemporaryFileCommandDelete:
		shouldRemoveTemporary = true

	case enumdef.TemporaryFileCommandSaveAsImage:
		shouldRemoveTemporary, resp, err = usecase.commandSaveAsImage(ctx, session)

	case enumdef.TemporaryFileCommandImageMetadata:
		shouldRemoveTemporary, resp, err = usecase.commandImageMetadata(ctx, session)

	case enumdef.TemporaryFileCommandChangeImageType:
		shouldRemoveTemporary, newContent, resp, err = usecase.commandChangeImageType(ctx, req.Metadata, session)
	default:
		err = xerror.Enrich(errordef.ErrRequestInvalid, "invalid command %s", req.Command)
	}

	if shouldRemoveTemporary {
		if err := usecase.fileStorageRepo.RemoveTemporary(ctx, session.Token); err != nil {
			xcontext.Logger(ctx).Warn("failed-to-remove-temporary-file", "err", err, "file", session.Token)
		}

		if err := usecase.fileSessionRepo.DeleteTemporarySession(ctx, session.Token); err != nil {
			xcontext.Logger(ctx).Warn("failed-to-remove-temporary-file-sesion", "err", err, "token", session.Token)
		}
	}

	if err != nil {
		return nil, err
	}

	if resp == nil {
		resp = &dto.CommandTemporaryFileResponse{}
	}

	if newContent != nil {
		newSession := usecase.fileDomain.NewTemporaryFileSession(session.UploadSessionInfo)
		metadata, err := usecase.fileStorageRepo.StoreTemporary(
			ctx, newSession.Token, newContent.Content, newContent.Size, newContent.Type)
		if err != nil {
			return nil, errordef.ErrServer.Hide(err, "failed-to-store-new-temporary-file")
		}

		newSession.FileHash = metadata.Hash
		if err := usecase.fileSessionRepo.SaveTemporarySession(ctx, newSession); err != nil {
			return nil, errordef.ErrServer.Hide(err, "failed-to-save-temporary-file-session")
		}

		resp.NextTemporaryFileToken = newSession.Token
	}

	return resp, nil
}

func (usecase *FileUsecase) commandSaveAsImage(
	ctx context.Context,
	session *domain.TemporaryFileSession,
) (bool, *dto.CommandTemporaryFileResponse, error) {
	ok, err := usecase.fileStorageRepo.ImageExists(ctx, session.FileHash)
	if err != nil {
		return false, nil, errordef.ErrServer.Hide(err, "failed-to-check-the-image-existence")
	}

	if !ok {
		if err := usecase.fileStorageRepo.ImageSave(ctx, session.FileHash, session.Token); err != nil {
			return false, nil, errordef.ErrServer.Hide(err, "failed-to-copy-from-session-image")
		}
	}

	return true, dto.NewTemporaryFileCommandSave(usecase.fileStorageRepo.ImageURL(session.FileHash)), nil
}

func (usecase *FileUsecase) commandImageMetadata(
	ctx context.Context,
	session *domain.TemporaryFileSession,
) (bool, *dto.CommandTemporaryFileResponse, error) {
	content, metadata, err := usecase.fileStorageRepo.GetTemporary(ctx, session.Token)
	if err != nil {
		return false, nil, errordef.ErrServer.Hide(err, "failed-to-get-temporary-file")
	}

	defer content.Close()

	if !mime.IsImage[metadata.Type] {
		return false, nil, xerror.Enrich(errordef.ErrRequestInvalid, "the file is not an image")
	}

	image, _, err := image.Decode(content)
	if err != nil {
		return true, nil, xerror.Enrich(errordef.ErrFileInvalidContent, "failed to decode file as an image").
			Hide(err, "failed-to-decode-the-image")
	}

	bounds := image.Bounds()
	result := dto.CommandImageMetadataResult{
		Type:       metadata.Type,
		Sha256Hash: metadata.Hash,
		Size:       metadata.Size,
		Width:      bounds.Dx(),
		Height:     bounds.Dy(),
	}

	resp, err := dto.NewCommandTemporaryFileRead(result)
	if err != nil {
		return false, nil, errordef.ErrServer.Hide(err, "failed-to-serialize-image-metadata-result")
	}

	return false, resp, nil
}

func (usecase *FileUsecase) commandChangeImageType(
	ctx context.Context,
	changeTo string,
	session *domain.TemporaryFileSession,
) (bool, *dto.FileContentWrapper, *dto.CommandTemporaryFileResponse, error) {
	if !mime.IsImage[changeTo] {
		return false, nil, nil, xerror.Enrich(errordef.ErrRequestInvalid, "invalid change mime type %s", changeTo)
	}

	content, metadata, err := usecase.fileStorageRepo.GetTemporary(ctx, session.Token)
	if err != nil {
		return false, nil, nil, errordef.ErrServer.Hide(err, "failed-to-get-temporary-file")
	}

	defer content.Close()

	if !mime.IsImage[metadata.Type] {
		return false, nil, nil, xerror.Enrich(errordef.ErrRequestInvalid, "the file is not an image")
	}

	if metadata.Type == changeTo {
		return false, nil, nil, nil
	}

	image, _, err := image.Decode(content)
	if err != nil {
		return true, nil, nil, xerror.Enrich(errordef.ErrFileInvalidContent, "failed to decode file as an image").
			Hide(err, "failed-to-decode-the-image")
	}

	buffer := bytes.NewBuffer(nil)
	switch changeTo {
	case mime.ImageJPEG:
		err = jpeg.Encode(buffer, image, nil)
	case mime.ImagePNG:
		err = png.Encode(buffer, image)
	}
	if err != nil {
		return false, nil, nil, xerror.Enrich(errordef.ErrFileInvalidContent, "cannot generate the new image").
			Hide(err, "failed-to-encode-the-new-content", "new", changeTo, "cur", metadata.Type)
	}

	newContent := &dto.FileContentWrapper{
		Content: io.NopCloser(buffer),
		Type:    changeTo,
		Size:    int64(buffer.Len()),
	}

	return true, newContent, nil, nil
}

func (*FileUsecase) translateValidateExternalPolicyError(err error, source string) error {
	if errors.Is(err, errordef.ErrForbidden) {
		return xerror.Enrich(errordef.ErrForbidden, "invalid policy token").
			Hide(err, "failed-to-validate-external-policy", "source", source)
	}

	return errordef.ErrServer.Hide(err, "failed-to-validate-external-policy-error", "source", source)
}

func (*FileUsecase) validatePolicy(
	req *dto.ValidatePolicyRequest,
	policy *domain.UploadPolicy,
) error {
	if !slices.Contains(policy.AllowedTypes, req.Type) {
		return xerror.Enrich(errordef.ErrFileMismatchedType, "unsupported file type")
	}

	if req.Size > policy.MaxSize {
		return xerror.Enrich(errordef.ErrFileMismatchedSize, "exceed the maxmium image size")
	}

	return nil
}
