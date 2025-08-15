package repository

import (
	"context"
	"fmt"
	"multi-tenant-service/package/structs"
	"time"
)

func (r *MessageRepository) 	GetMessages(ctx context.Context, req structs.RequestGetMessage, cursorTime time.Time) ([]structs.Message,error){
	// Build query
	var query string
	var args []interface{}
	var messages []structs.Message

	if req.Cursor != nil && *req.Cursor != "" {
		query = `
			SELECT id, tenant_id, payload, created_at
			FROM messages
			WHERE tenant_id = $1 AND created_at > $2
			ORDER BY created_at ASC
			LIMIT $3
		`
		args = []interface{}{req.TenantID, cursorTime, req.Limit + 1}
	} else {
		query = `
			SELECT id, tenant_id, payload, created_at
			FROM messages
			WHERE tenant_id = $1
			ORDER BY created_at ASC
			LIMIT $2
		`
		args = []interface{}{req.TenantID, req.Limit + 1}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var msg structs.Message
		if err := rows.Scan(&msg.ID, &msg.TenantID, &msg.Payload, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}