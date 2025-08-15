package structs

type CreateTenantRequest struct {
	Name              string `json:"name" binding:"required"`
	ConcurrencyConfig int    `json:"concurrency_config"`
}