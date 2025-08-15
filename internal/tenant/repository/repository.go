package repository

import (
	"context"
	"multi-tenant-service/package/connection/database"
	"multi-tenant-service/package/structs"
)

type TenantRepository struct {
	db *database.DB
}

type ITenantRepository interface {
	CreateTenant(ctx context.Context, tenant structs.Tenant) error
	DropTenantPartition(ctx context.Context, tenantID string) error
	DeleteTenant(ctx context.Context, tenantID string) error
	CreateTenantPartition(tenantID string) error
	UpdateTenantConcurrency(ctx context.Context, tenantID string, workers int) error
	GetTenant(ctx context.Context, tenantID string) (*structs.Tenant, error)
}


func NewTenantRepository(db *database.DB) ITenantRepository  {
	return &TenantRepository{
		db: db,
	}
}