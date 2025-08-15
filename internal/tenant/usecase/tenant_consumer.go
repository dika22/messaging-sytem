package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"multi-tenant-service/package/structs"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (tu *TenantUsecase) startTenantConsumer(ctx context.Context, tenantID string, workers int) error {
	channelName := fmt.Sprintf("tenant_%s", tenantID)
	queueName := fmt.Sprintf("tenant_%s_queue", tenantID)

	// Create channel
	ch, err := tu.mqClient.CreateChannel(channelName)
	if err != nil {
		return err
	}

	// Declare queue
	_, err = tu.mqClient.DeclareQueue(ch, queueName)
	if err != nil {
		return err
	}

	// Create consumer
	consumer := &TenantConsumer{
		Channel:    ch,
		StopChan:   make(chan bool),
		Workers:    int64(workers),
		WorkerPool: make(chan struct{}, workers),
	}

	// Initialize worker pool
	for i := 0; i < workers; i++ {
		consumer.WorkerPool <- struct{}{}
	}

	tu.mu.Lock()
	tu.consumers[tenantID] = consumer
	tu.mu.Unlock()

	// Start consuming messages
	go tu.consumeMessages(ctx, tenantID, consumer, queueName)

	return nil
}

func (tm *TenantUsecase) consumeMessages(ctx context.Context, tenantID string, consumer *TenantConsumer, queueName string) {
	msgs, err := consumer.Channel.Consume(
		queueName,
		fmt.Sprintf("consumer_%s", tenantID),
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Printf("Failed to start consuming for tenant %s: %v", tenantID, err)
		return
	}

	for {
		select {
		case <-consumer.StopChan:
			log.Printf("Stopping consumer for tenant %s", tenantID)
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("Message channel closed for tenant %s", tenantID)
				return
			}

			// Get worker from pool
			<-consumer.WorkerPool

			// Process message in goroutine
			go func(msg amqp.Delivery) {
				defer func() {
					consumer.WorkerPool <- struct{}{} // Return worker to pool
				}()

				if err := tm.processMessage(ctx, tenantID, msg); err != nil {
					log.Printf("Failed to process message for tenant %s: %v", tenantID, err)
					msg.Nack(false, true) // Requeue message
				} else {
					msg.Ack(false)
				}
			}(msg)
		}
	}
}

func (tu *TenantUsecase) processMessage(ctx context.Context, tenantID string, msg amqp.Delivery) error {
	// Parse message
	var messageReq structs.CreateMessageRequest
	if err := json.Unmarshal(msg.Body, &messageReq); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	tu.msgRepo.InsertMessage(ctx, messageReq)
	log.Printf("Processed message for tenant %s", tenantID)
	return nil
}

func (tm *TenantUsecase) Shutdown(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	log.Println("Shutting down tenant consumers...")

	for tenantID, consumer := range tm.consumers {
		log.Printf("Stopping consumer for tenant %s", tenantID)
		close(consumer.StopChan)
		tm.mqClient.CloseChannel(fmt.Sprintf("tenant_%s", tenantID))
	}

	tm.consumers = make(map[string]*TenantConsumer)
	return nil
}