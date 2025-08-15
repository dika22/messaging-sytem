package delivery

import (
	"multi-tenant-service/internal/tenant/usecase"
	"multi-tenant-service/package/response"
	"multi-tenant-service/package/structs"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TenantHTTPHandler struct {
	tenantUsecase usecase.ITenantUsecase
}


// CreateTenant godoc
// @Summary Create a new tenant
// @Description Create a new tenant with dedicated queue and consumer
// @Tags tenants
// @Accept json
// @Produce json
// @Param tenant body structs.CreateTenantRequest true "Tenant data"
// @Success 201 {object} structs.Tenant
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants [post]
func (h *TenantHTTPHandler) CreateTenant(c echo.Context) error {
	ctx := c.Request().Context()
	var req structs.CreateTenantRequest
	if err := c.Bind(&req); err != nil {
		return response.JSONResponse(c, http.StatusBadRequest, false, err.Error(), nil)
	}

	tenant, err := h.tenantUsecase.CreateTenant(ctx, req)
	if err != nil {
		return response.JSONResponse(c, http.StatusInternalServerError, false, err.Error(), nil)
	}
	return response.JSONResponse(c, http.StatusCreated, true, "Tenant created successfully", tenant)
}


// DeleteTenant godoc
// @Summary Delete tenant
// @Description Delete tenant and stop its consumer
// @Tags tenants
// @Param id path string true "Tenant ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants/{id} [delete]
func (h *TenantHTTPHandler) DeleteTenant(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID := c.Param("id")
	if _, err := uuid.Parse(tenantID); err != nil {
		return response.JSONResponse(c, http.StatusBadRequest, false, "Invalid tenant ID", nil)
	}

	if err := h.tenantUsecase.DeleteTenant(ctx, tenantID); err != nil {
		return response.JSONResponse(c, http.StatusInternalServerError, false, err.Error(), nil)
	}
	return response.JSONResponse(c, http.StatusNoContent, true, "Tenant deleted successfully", nil)
}

// UpdateConcurrency godoc
// @Summary Update tenant concurrency
// @Description Update the number of concurrent workers for a tenant
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param config body structs.UpdateConcurrencyRequest true "Concurrency config"
// @Success 200 {object} structs.Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tenants/{id}/config/concurrency [put]
func (h *TenantHTTPHandler) UpdateConcurrency(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID := c.Param("id")
	if _, err := uuid.Parse(tenantID); err != nil {
		response.JSONResponse(c, http.StatusBadRequest, false, "Invalid tenant ID", nil)
	}

	var req structs.UpdateConcurrencyRequest
	if err := c.Bind(&req); err != nil {
		return response.JSONResponse(c, http.StatusBadRequest, false, err.Error(), nil)
	}

	if err := h.tenantUsecase.UpdateTenantConcurrency(ctx, tenantID, req.Workers); err != nil {
		if err.Error() == "tenant not found" {
			return response.JSONResponse(c, http.StatusNotFound, false, err.Error(), nil)
		}
		return response.JSONResponse(c, http.StatusInternalServerError, false, err.Error(), nil)
	}

	return response.JSONResponse(c, http.StatusOK, true, "Concurrency updated successfully", nil)
}


func NewTenantHTTPHandler(r *echo.Group, tenantUsecase usecase.ITenantUsecase)  {
	h := &TenantHTTPHandler{
		tenantUsecase: tenantUsecase,
	}
	r.POST("/tenants", h.CreateTenant).Name = "CreateTenant"
	r.DELETE("/tenants/:id", h.DeleteTenant).Name = "DeleteTenant"
	r.PUT("/tenants/:id/config/concurrency", h.UpdateConcurrency).Name = "UpdateConcurrency"
}