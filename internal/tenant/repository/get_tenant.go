package repository

import (
	"context"
	"database/sql"
	"fmt"
	"multi-tenant-service/package/structs"
)

func (r TenantRepository) GetTenant(ctx context.Context, tenantID string) (*structs.Tenant, error) {
	tenant := &structs.Tenant{}
	query := `
		SELECT id, name, concurrency_config, created_at, updated_at
		FROM tenants WHERE id = $1
	`
	err :=r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&tenant.ID, &tenant.Name, &tenant.ConcurrencyConfig,
		&tenant.CreatedAt, &tenant.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}