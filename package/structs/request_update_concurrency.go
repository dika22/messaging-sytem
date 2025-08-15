package structs

type UpdateConcurrencyRequest struct {
	Workers int `json:"workers" binding:"required,min=1"`
}