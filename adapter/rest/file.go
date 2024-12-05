package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/todennus/file-service/adapter/abstraction"
	"github.com/todennus/file-service/adapter/rest/dto"
	"github.com/todennus/shared/errordef"
	"github.com/todennus/shared/middleware"
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
	r.Get("/token/{ownership_id}", middleware.RequireAuthentication(a.RetrieveFileToken()))
	r.Post("/", a.Upload()) // Has already required authentication in the handler.
}

// @Summary Upload file.
// @Description Use an `upload_token` to upload the file. This API also returns a `file_token`.
// @Tags File
// @Accept multipart/form-data
// @Produce json
// @Param upload_token formData string true "upload token"
// @Param file formData file true "Upload file"
// @Success 201 {object} response.SwaggerSuccessResponse[dto.UploadResponse] "Upload successfully"
// @Failure 400 {object} response.SwaggerBadRequestErrorResponse "Bad request"
// @Failure 403 {object} response.SwaggerForbiddenErrorResponse "Forbidden"
// @Router /files [post]
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
		// it's better to close the connection when an error occurs to avoid
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

// @Summary Retrieve file token.
// @Description Use an `ownership_id` to retrieve a file token. This token can be used to interact with file in other APIs.
// @Tags File
// @Produce json
// @Param user body dto.RetrieveFileTokenRequest true "Retrieve policy request"
// @Success 201 {object} response.SwaggerSuccessResponse[dto.RetrieveFileTokenResponse] "Successfully retrieve the file token"
// @Failure 400 {object} response.SwaggerBadRequestErrorResponse "Bad request"
// @Router /files/token/{ownership_id} [get]
func (a *FileAdapter) RetrieveFileToken() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		req, err := xhttp.ParseHTTPRequest[dto.RetrieveFileTokenRequest](r)
		if err != nil {
			response.RESTWriteAndLogInvalidRequestError(ctx, w, err)
			return
		}

		resp, err := a.fileUsecase.RetrieveFileToken(ctx, req.To())
		response.NewRESTResponseHandler(ctx, dto.NewRetrieveFileTokenResponse(resp), err).
			Map(http.StatusBadRequest, errordef.ErrRequestInvalid).
			WriteHTTPResponse(ctx, w)
	}
}
