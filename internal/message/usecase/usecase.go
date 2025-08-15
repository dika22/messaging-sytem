package usecase

import (
	"context"
	"multi-tenant-service/internal/message/repository"
	repoTenant "multi-tenant-service/internal/tenant/repository"

	rabbitmq "multi-tenant-service/package/rabbit-mq"
	"multi-tenant-service/package/structs"
)

type MessageUsecase struct {
	repository repository.IMessageRepository
	repoTenant repoTenant.ITenantRepository
	mqClient *rabbitmq.Client
}

type IMessageUsecase interface {
	GetMessages(ctx context.Context, req structs.RequestGetMessage) (*structs.MessageResponse, error)
	PublishMessage(ctx context.Context, req structs.CreateMessageRequest) error
}



func NewMessageUsecase(messgeRepo repository.IMessageRepository, 
	repoTenant repoTenant.ITenantRepository,
	mqClient *rabbitmq.Client) IMessageUsecase {
	return &MessageUsecase{
		repository: messgeRepo,
		repoTenant: repoTenant,
		mqClient: mqClient,
	}
	
}