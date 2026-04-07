package handlers

// errorDetail holds the machine-readable code and human-readable message.
type errorDetail struct {
	Code    string `json:"code"    example:"FORM_NOT_FOUND"`
	Message string `json:"message" example:"form not found"`
}

// errorResponse is the standard error envelope returned by all endpoints.
type errorResponse struct {
	Error errorDetail `json:"error"`
}
