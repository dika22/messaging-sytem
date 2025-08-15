package usecase

import (
	"context"
	"fmt"
	"log"
	"multi-tenant-service/package/structs"

	"github.com/google/uuid"
)

func (tu *TenantUsecase) CreateTenant(ctx context.Context, req structs.CreateTenantRequest) (*structs.Tenant, error)   {
	tenantID := uuid.New()
	// Set default concurrency if not provided
	if req.ConcurrencyConfig == 0 {
		req.ConcurrencyConfig = 3
	}

	// Insert tenant into database
	tenant := &structs.Tenant{
		ID:                tenantID,
		Name:              req.Name,
		ConcurrencyConfig: req.ConcurrencyConfig,
	}

	if err := tu.repository.CreateTenant(ctx, *tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create partition table
	if err := tu.repository.CreateTenantPartition(tenantID.String()); err != nil {
		log.Printf("Warning: failed to create partition for tenant %s: %v", tenantID.String(), err)
		return nil, fmt.Errorf("failed to create partition: %w", err)
	}

	// Create RabbitMQ queue and start consumer
	if err := tu.startTenantConsumer(ctx, tenantID.String(), req.ConcurrencyConfig); err != nil {
		return nil, fmt.Errorf("failed to start consumer: %w", err)
	}

	return tenant, nil
}