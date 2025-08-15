package structs

import "github.com/google/uuid"

type RequestGetMessage struct {
	TenantID   uuid.UUID `json:"tenant_id" binding:"required"`
	Cursor     *string   `json:"cursor"`
	CursorTime *string `json:"cursor_time"`
	Limit      int       `json:"limit"`
}