package repository

import (
	"context"
	"multi-tenant-service/package/structs"
)

func (r TenantRepository) CreateTenant(ctx context.Context, tenant structs.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, concurrency_config)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, tenant.ID, tenant.Name, tenant.ConcurrencyConfig).
		Scan(&tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}