package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/todennus/file-service/adapter/abstraction"
	"github.com/todennus/file-service/adapter/rest/dto"
	"github.com/todennus/shared/errordef"
	"github.com/todennus/shared/response"
	"github.com/todennus/x/xhttp"
)

type FileAdapter struct {
	fileUsecase abstraction.FileUsecase
}

func NewFileAdapter(fileUsecase abstraction.FileUsecase) *FileAdapter {
	return &FileAdapter{fileUsecase: fileUsecase}
}

func (a *FileAdapter) Router(r chi.Router) {
	r.Post("/policy/validate", a.ValidatePolicy())
	r.Post("/upload", a.Upload())
}

// @Summary Validate file metadata.
// @Description Use a `policy_token` to validate the file metadata against the policy. If the validation is success, get an `upload_token` to upload the file later.
// @Tags File
// @Accept json
// @Produce json
// @Param user body dto.ValidatePolicyRequest true "Validate policy request"
// @Success 201 {object} response.SwaggerSuccessResponse[dto.ValidatePolicyResponse] "Successfully validate the file metadata"
// @Failure 400 {object} response.SwaggerBadRequestErrorResponse "Bad request"
// @Failure 403 {object} response.SwaggerForbiddenErrorResponse "Forbidden"
// @Router /files/policy/validate [post]
func (a *FileAdapter) ValidatePolicy() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		request, err := xhttp.ParseHTTPRequest[dto.ValidatePolicyRequest](r)
		if err != nil {
			response.RESTWriteAndLogInvalidRequestError(ctx, w, err)
			return
		}

		resp, err := a.fileUsecase.ValidatePolicy(ctx, request.To())
		response.NewRESTResponseHandler(ctx, dto.NewValidatePolicyResponse(resp), err).
			Map(http.StatusBadRequest,
				errordef.ErrRequestInvalid,
				errordef.ErrFileMismatchedType,
				errordef.ErrFileMismatchedSize,
			).
			Map(http.StatusForbidden, errordef.ErrForbidden).
			WriteHTTPResponse(ctx, w)
	}
}

// @Summary Upload file.
// @Description Use an `upload_token` to upload the file. This file will be stored in temporary storage. Return a `temporary_file_token`.
// @Tags File
// @Accept multipart/form-data
// @Produce json
// @Param upload_token formData string true "upload token"
// @Param file formData file true "Upload file"
// @Success 201 {object} response.SwaggerSuccessResponse[dto.UploadResponse] "Upload successfully"
// @Failure 400 {object} response.SwaggerBadRequestErrorResponse "Bad request"
// @Failure 403 {object} response.SwaggerForbiddenErrorResponse "Forbidden"
// @Router /files/upload [post]
func (a *FileAdapter) Upload() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		request, err := xhttp.ParseHTTPRequest[dto.UploadRequest](r)
		if err != nil {
			response.RESTWriteAndLogInvalidRequestError(ctx, w, err)
			return
		}

		ucreq, err := request.To()
		if err != nil {
			response.RESTWriteAndLogInvalidRequestError(ctx, w, err)
			return
		}

		resp, err := a.fileUsecase.Upload(ctx, ucreq)

		// If the usecase rejects the uploaded file, the remaining request body
		// would still need to be read into memory before reusing the
		// connection, which could introduce unnecessary latency. Therefore,
		// it’s better to close the connection when an error occurs to avoid
		// this overhead.
		if err != nil {
			w.Header().Set("Connection", "close")
		}

		response.NewRESTResponseHandler(ctx, dto.NewUploadResponse(resp), err).
			Map(http.StatusBadRequest,
				errordef.ErrRequestInvalid,
				errordef.ErrFileInvalidContent,
				errordef.ErrFileMismatchedType,
				errordef.ErrFileMismatchedSize,
			).
			Map(http.StatusForbidden, errordef.ErrForbidden).
			WithDefaultCode(http.StatusCreated).
			WriteHTTPResponse(ctx, w)
	}
}
