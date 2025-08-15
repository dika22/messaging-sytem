package usecase

import (
	"context"
	"sync/atomic"
)

func (tu *TenantUsecase) UpdateTenantConcurrency(ctx context.Context, tenantID string, workers int) error {
	tu.mu.Lock()
	defer tu.mu.Unlock()

	if err := tu.repository.UpdateTenantConcurrency(ctx, tenantID, workers); err != nil {
		return err
	}

	// Update consumer if exists
	if consumer, exists := tu.consumers[tenantID]; exists {
		atomic.StoreInt64(&consumer.Workers, int64(workers))
		consumer.WorkerPool = make(chan struct{}, workers)
		for i := 0; i < workers; i++ {
			consumer.WorkerPool <- struct{}{}
		}
	}
	return nil
}