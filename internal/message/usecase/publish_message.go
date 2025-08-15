package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"multi-tenant-service/package/structs"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (mu *MessageUsecase) PublishMessage(ctx context.Context, req structs.CreateMessageRequest)  error {
	tenant, err := mu.repoTenant.GetTenant(ctx, req.TenantID.String())
	if err != nil {
		return err
	}

	if tenant.ID == uuid.Nil {
		return fmt.Errorf("tenant not found")
	}

	queueName := fmt.Sprintf("tenant_%s_queue", req.TenantID.String())
	// Create channel for publishing
	ch, err := mu.mqClient.CreateChannel("publisher")
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	// Ensure queue exists
	_, err = mu.mqClient.DeclareQueue(ch, queueName)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Marshal message
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish message
	err = ch.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}