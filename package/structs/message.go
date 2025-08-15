package structs

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	TenantID  uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Payload   map[string]interface{} `json:"payload" db:"payload"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

type CreateMessageRequest struct {
	TenantID uuid.UUID       `json:"tenant_id" binding:"required"`
	Payload  map[string]interface{} `json:"payload" binding:"required"`
}

type MessageResponse struct {
	Data       []Message `json:"data"`
	NextCursor *string   `json:"next_cursor,omitempty"`
}