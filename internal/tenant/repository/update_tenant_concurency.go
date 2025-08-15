package repository

import (
	"context"
	"fmt"
)

func (r TenantRepository) UpdateTenantConcurrency(ctx context.Context, tenantID string, workers int) error {
	// Update database
	_, err := r.db.ExecContext(ctx, 
		"UPDATE tenants SET concurrency_config = $1, updated_at = NOW() WHERE id = $2",
		workers, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update concurrency: %w", err)
	}
	return  nil
}