package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)
func (r *MessageRepository) GetMessageCount(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM messages WHERE tenant_id = $1"
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get message count: %w", err)
	}
	return  count, nil
}