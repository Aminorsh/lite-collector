package handlers

import "time"

// errorDetail holds the machine-readable code and human-readable message.
type errorDetail struct {
	Code    string `json:"code"    example:"FORM_NOT_FOUND"`
	Message string `json:"message" example:"form not found"`
}

// errorResponse is the standard error envelope returned by all endpoints.
type errorResponse struct {
	Error errorDetail `json:"error"`
}

type jobStatusResponse struct {
	ID         uint64     `json:"id"          example:"1"`
	JobType    string     `json:"job_type"    example:"detect_anomaly"`
	Status     int8       `json:"status"      example:"0"`
	Input      string     `json:"input,omitempty"`
	Output     string     `json:"output,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// pendingJobItem is the trimmed shape used by the banner — no Input/Output blobs.
type pendingJobItem struct {
	ID         uint64     `json:"id"`
	JobType    string     `json:"job_type"`
	Status     int8       `json:"status"`
	FormID     *uint64    `json:"form_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type pendingJobsResponse struct {
	Jobs []pendingJobItem `json:"jobs"`
}
