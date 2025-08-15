package usecase

import (
	"context"
	"fmt"
	"multi-tenant-service/package/structs"
)

func (tu *TenantUsecase) GetTenant(ctx context.Context, tenantID string) (*structs.Tenant, error) {
	tenant, err := tu.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return tenant, nil
}
