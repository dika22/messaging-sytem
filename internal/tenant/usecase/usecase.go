package usecase

import (
	"context"
	"multi-tenant-service/internal/tenant/repository"
	rabbitmq "multi-tenant-service/package/rabbit-mq"
	"multi-tenant-service/package/structs"
	"sync"

	rm "multi-tenant-service/internal/message/repository"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TenantUsecase struct {
	repository repository.ITenantRepository
	msgRepo    rm.IMessageRepository
	mqClient *rabbitmq.Client
	consumers map[string]*TenantConsumer
	mu        sync.RWMutex
}

type TenantConsumer struct {
	Channel    *amqp.Channel
	StopChan   chan bool
	Workers    int64
	WorkerPool chan struct{}
}


type ITenantUsecase interface {
	CreateTenant(ctx context.Context, req structs.CreateTenantRequest) (*structs.Tenant, error) 
	DeleteTenant(ctx context.Context, tenantID string) error
	GetTenant(ctx context.Context, tenantID string) (*structs.Tenant, error)
	UpdateTenantConcurrency(ctx context.Context, tenantID string, workers int) error
	// ListTenant(req structs.RequestListTenant) (structs.ResponseListTenant, error)
}


func NewTenantUsecase(tenantRepo repository.ITenantRepository,
	msgRepo rm.IMessageRepository, mqClient *rabbitmq.Client) ITenantUsecase {
	return &TenantUsecase{
		repository: tenantRepo,
		msgRepo   : msgRepo,
		mqClient  : mqClient,
		consumers: make(map[string]*TenantConsumer),
	}
}