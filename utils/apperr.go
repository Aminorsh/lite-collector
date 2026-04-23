package utils

import (
	"errors"
	"net/http"
)

// AppError is a typed application error carrying an HTTP status, a machine-readable
// code for the frontend, and a human-readable message.
type AppError struct {
	HTTPStatus int
	Code       string
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}

// Predefined errors — add new ones here as the app grows.
var (
	ErrBadRequest = &AppError{http.StatusBadRequest, "BAD_REQUEST", "invalid request"}
	ErrUnauthorized = &AppError{http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized"}
	ErrForbidden    = &AppError{http.StatusForbidden, "FORBIDDEN", "forbidden"}
	ErrNotFound     = &AppError{http.StatusNotFound, "NOT_FOUND", "resource not found"}
	ErrInternal     = &AppError{http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"}

	// Auth
	ErrLoginFailed       = &AppError{http.StatusInternalServerError, "LOGIN_FAILED", "login failed"}
	ErrWxCodeExchangeFail = &AppError{http.StatusBadGateway, "WX_CODE_EXCHANGE_FAILED", "failed to exchange WeChat code for session"}

	// Forms
	ErrFormNotFound   = &AppError{http.StatusNotFound, "FORM_NOT_FOUND", "form not found"}
	ErrFormForbidden  = &AppError{http.StatusForbidden, "FORM_FORBIDDEN", "you do not own this form"}
	ErrFormCreateFail  = &AppError{http.StatusInternalServerError, "FORM_CREATE_FAILED", "failed to create form"}
	ErrFormUpdateFail  = &AppError{http.StatusInternalServerError, "FORM_UPDATE_FAILED", "failed to update form"}
	ErrFormPublishFail = &AppError{http.StatusInternalServerError, "FORM_PUBLISH_FAILED", "failed to publish form"}
	ErrFormArchiveFail = &AppError{http.StatusInternalServerError, "FORM_ARCHIVE_FAILED", "failed to archive form"}
	ErrFormNotPublished = &AppError{http.StatusForbidden, "FORM_NOT_PUBLISHED", "form is not published"}

	// Submissions
	ErrSubmissionNotFound   = &AppError{http.StatusNotFound, "SUBMISSION_NOT_FOUND", "submission not found"}
	ErrSubmissionCreateFail = &AppError{http.StatusInternalServerError, "SUBMISSION_CREATE_FAILED", "failed to create submission"}

	// AI
	ErrAINotConfigured = &AppError{http.StatusServiceUnavailable, "AI_NOT_CONFIGURED", "AI service is not configured (DEEPSEEK_API_KEY not set)"}
	ErrAIGenerateFail  = &AppError{http.StatusBadGateway, "AI_GENERATE_FAILED", "AI failed to generate form schema"}

	// PDF
	ErrJobNotCompleted = &AppError{http.StatusConflict, "JOB_NOT_COMPLETED", "job has not finished successfully yet"}
	ErrJobNotReportable = &AppError{http.StatusBadRequest, "JOB_NOT_REPORTABLE", "PDF export is only supported for generate_report jobs"}
	ErrPDFGenerateFail  = &AppError{http.StatusInternalServerError, "PDF_GENERATE_FAILED", "failed to render PDF"}
	ErrPDFNotAvailable  = &AppError{http.StatusServiceUnavailable, "PDF_NOT_AVAILABLE", "PDF rendering is not available on this server (chromium missing)"}
)

// AsAppError unwraps err into an *AppError. If err is not an *AppError,
// it returns ErrInternal so handlers always have a typed error to work with.
func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return ErrInternal
}
