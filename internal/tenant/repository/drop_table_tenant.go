package repository

import (
	"context"
	"fmt"
)

func (r TenantRepository) DropTenantPartition(ctx context.Context, tenantID string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS messages_tenant_%s", tenantID)
	_, err := r.db.Exec(query)
	return err
}