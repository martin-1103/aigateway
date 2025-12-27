package manager

import (
	"context"
	"log"
	"time"

	"aigateway-backend/models"
)

const DefaultReconcileInterval = 5 * time.Minute

// StartPeriodicReconcile starts background account reconciliation
func (m *Manager) StartPeriodicReconcile(ctx context.Context, interval time.Duration, providerIDs []string) {
	if interval <= 0 {
		interval = DefaultReconcileInterval
	}

	// Cancel previous loop if exists
	if m.reconcileCancel != nil {
		m.reconcileCancel()
	}

	reconcileCtx, cancel := context.WithCancel(ctx)
	m.reconcileCancel = cancel

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Defer first run by 30 seconds to allow server startup
		time.Sleep(30 * time.Second)
		m.reconcileAccounts(reconcileCtx, providerIDs)

		for {
			select {
			case <-reconcileCtx.Done():
				return
			case <-ticker.C:
				m.reconcileAccounts(reconcileCtx, providerIDs)
			}
		}
	}()
}

// StopPeriodicReconcile stops background reconciliation
func (m *Manager) StopPeriodicReconcile() {
	if m.reconcileCancel != nil {
		m.reconcileCancel()
		m.reconcileCancel = nil
	}
}

// reconcileAccounts syncs DB state to in-memory map
func (m *Manager) reconcileAccounts(ctx context.Context, providerIDs []string) {
	startTime := time.Now()

	for _, providerID := range providerIDs {
		// Query DB for all active accounts
		dbAccounts, err := m.accountRepo.GetActiveByProvider(providerID)
		if err != nil {
			log.Printf("[AuthManager] Reconcile failed for %s: %v", providerID, err)
			continue
		}

		// Build map of DB account IDs
		dbAccountIDs := make(map[string]*models.Account)
		for _, acc := range dbAccounts {
			dbAccountIDs[acc.ID] = acc
		}

		// Acquire write lock for reconciliation
		m.mu.Lock()

		// Track changes
		var added, removed, unchanged int

		// Find accounts in memory but not in DB (deleted)
		for id, accState := range m.accounts {
			if accState.Account.ProviderID != providerID {
				continue
			}
			if _, exists := dbAccountIDs[id]; !exists {
				delete(m.accounts, id)
				removed++
				log.Printf("[AuthManager] Reconcile: Removed account %s (deleted from DB)", id)
			} else {
				unchanged++
			}
		}

		// Find accounts in DB but not in memory (missing)
		for id, acc := range dbAccountIDs {
			if _, exists := m.accounts[id]; !exists {
				m.accounts[id] = NewAccountState(acc)
				added++
				log.Printf("[AuthManager] Reconcile: Added account %s (missing from memory)", id)
			}
		}

		m.mu.Unlock()

		// Log summary
		if added > 0 || removed > 0 {
			log.Printf("[AuthManager] Reconcile %s: +%d -%d =%d (took %v)",
				providerID, added, removed, unchanged, time.Since(startTime))
		}
	}
}
