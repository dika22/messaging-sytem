package usecase

import (
	"context"
	"fmt"
	"log"
)


func (tu *TenantUsecase) DeleteTenant(ctx context.Context, tenantID string) error {
	tu.mu.Lock()
	defer tu.mu.Unlock()

	// Stop consumer
	if consumer, exists := tu.consumers[tenantID]; exists {
		close(consumer.StopChan)
		tu.mqClient.CloseChannel(fmt.Sprintf("tenant_%s", tenantID))
		delete(tu.consumers, tenantID)
	}

	// Delete from database
	if err := tu.repository.DeleteTenant(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Drop partition table
	if err := tu.repository.DropTenantPartition(ctx, tenantID); err != nil {
		log.Printf("Warning: failed to drop partition for tenant %s: %v", tenantID, err)
	}

	return nil
}