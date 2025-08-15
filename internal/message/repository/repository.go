package repository

import (
	"context"
	"multi-tenant-service/package/connection/database"
	"multi-tenant-service/package/structs"
	"time"

	"github.com/google/uuid"
)

type MessageRepository struct {
	db *database.DB
}

type IMessageRepository interface {
	PublishMessage(ctx context.Context, req structs.CreateMessageRequest) error
	GetMessages(ctx context.Context, req structs.RequestGetMessage, cursorTime time.Time) ([]structs.Message,error)
	GetMessageCount(ctx context.Context, tenantID uuid.UUID) (int, error)
	InsertMessage(ctx context.Context, req structs.CreateMessageRequest) error
}


func NewMessageRepository(db *database.DB) IMessageRepository {
	return &MessageRepository{
		db: db,
	}
}