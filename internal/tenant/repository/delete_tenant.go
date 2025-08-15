package repository

import "context"

func (r *TenantRepository) DeleteTenant(ctx context.Context, tenantID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tenants WHERE id = $1", tenantID)
	if err != nil {
		return err
	}
	return  nil
}