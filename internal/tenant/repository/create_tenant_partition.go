package repository

import (
	"fmt"
	"strings"
)

func (r TenantRepository) CreateTenantPartition(tenantID string) error {
	unixID := strings.Replace(tenantID, "-", "", -1)
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS messages_tenant_%s 
		PARTITION OF messages 
		FOR VALUES IN ('%s')
	`, unixID, tenantID)

	_, err := r.db.Exec(query)
	return err
}
