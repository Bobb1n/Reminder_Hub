package response

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateIntegrationResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}