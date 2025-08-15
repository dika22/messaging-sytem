package structs
type Response struct {
	Result interface{} `json:"result"`
	Message string `json:"message"`
	Status bool `json:"status"`
	StatusCode int64 `json:"status_code"`
}