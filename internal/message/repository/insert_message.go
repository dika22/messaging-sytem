package repository

import (
	"context"
	"fmt"
	"log"
	"multi-tenant-service/package/structs"
)

func (r MessageRepository) InsertMessage(ctx context.Context, req structs.CreateMessageRequest) error {
	// Store message in database
	query := `
		INSERT INTO messages (tenant_id, payload)
		VALUES ($1, $2)
	`
	_, err := r.db.ExecContext(ctx, query, req.TenantID, req.Payload)
	if err != nil {
		return fmt.Errorf("failed to store message: %w", err)
	}

	log.Printf("Processed message for tenant %s", req.TenantID)
	return nil
}