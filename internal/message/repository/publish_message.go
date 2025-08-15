package repository

import (
	"context"
	"multi-tenant-service/package/structs"
)

func (r *MessageRepository) PublishMessage(ctx context.Context, payload structs.CreateMessageRequest) error{
	return  nil
}