package usecase

import (
	"context"
	"fmt"
	"multi-tenant-service/package/structs"
	"time"

	"github.com/google/uuid"
)

func (mu *MessageUsecase) GetMessages(ctx context.Context, req structs.RequestGetMessage) (*structs.MessageResponse, error) {	
	var nextCursor *string
	// Parse cursor if provided
	var cursorTime time.Time
	if req.Cursor != nil && *req.Cursor != "" {
		parsedTime, err := time.Parse(time.RFC3339, *req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor format: %w", err)
		}
		cursorTime = parsedTime
	}

	// Get messages
	messages, err := mu.repository.GetMessages(ctx, req, cursorTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	
	// Check if there are more results
	if len(messages) > req.Limit {
		// Remove extra message and set next cursor
		messages = messages[:req.Limit]
		lastMsg := messages[len(messages)-1]
		cursorStr := lastMsg.CreatedAt.Format(time.RFC3339)
		nextCursor = &cursorStr
	}

	return &structs.MessageResponse{
		Data:       messages,
		NextCursor: nextCursor,
	}, nil
}

func (mu *MessageUsecase) GetMessageCount(ctx context.Context, tenantID uuid.UUID) (int, error) {
	count, err := mu.repository.GetMessageCount(ctx, tenantID)
	if err != nil {
		return 0, err
	}
	return count, nil
}