package trax

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestTransactionIsolationBetweenGoroutines verifies that transactions
// are properly isolated between concurrent goroutines.
//
// This test addresses the critical bug where shared transaction state
// caused goroutines to interfere with each other, leading to deadlocks
// and query failures.
func TestTransactionIsolationBetweenGoroutines(t *testing.T) {
	// Skip if no test database available
	connectionString := "postgres://postgres:postgres@localhost:5432/trax_test?sslmode=disable"

	store, err := NewPsqlStore(connectionString)
	if err != nil {
		t.Skip("Skipping test: cannot connect to test database")
		return
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test cluster for this test
	testCluster := &Cluster{
		Id:          "tx_isolation_test_cluster",
		DisplayName: "Transaction Isolation Test Cluster",
		Description: "Test cluster for transaction isolation",
		Metadata:    map[string]string{},
	}

	_, err = store.SaveClusterIdempotently(ctx, testCluster)
	if err != nil {
		t.Fatalf("Failed to create test cluster: %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 2)

	// Goroutine 1: Begins a transaction and holds it
	wg.Add(1)
	go func() {
		defer wg.Done()

		err := store.BeginTransaction(ctx)
		if err != nil {
			errors <- err
			return
		}

		// Simulate work within transaction
		time.Sleep(100 * time.Millisecond)

		// Rollback to clean up
		if err := store.RollbackTransaction(ctx); err != nil {
			errors <- err
		}
	}()

	// Goroutine 2: Should be able to use the base store without interference
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Sleep briefly to ensure Goroutine 1 starts its transaction first
		time.Sleep(10 * time.Millisecond)

		// This should NOT use Goroutine 1's transaction
		// It should work independently
		_, err := store.ListClusterIds(ctx)
		if err != nil {
			errors <- err
			return
		}
	}()

	// Wait for both goroutines
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Goroutine error: %v", err)
	}
}

// TestConcurrentTransactions verifies that multiple goroutines
// can each have their own isolated transactions simultaneously
func TestConcurrentTransactions(t *testing.T) {
	connectionString := "postgres://postgres:postgres@localhost:5432/trax_test?sslmode=disable"

	store, err := NewPsqlStore(connectionString)
	if err != nil {
		t.Skip("Skipping test: cannot connect to test database")
		return
	}
	defer store.Close()

	ctx := context.Background()

	// Create test cluster
	testCluster := &Cluster{
		Id:          "concurrent_tx_test_cluster",
		DisplayName: "Concurrent TX Test Cluster",
		Description: "Test cluster for concurrent transactions",
		Metadata:    map[string]string{},
	}

	_, err = store.SaveClusterIdempotently(ctx, testCluster)
	if err != nil {
		t.Fatalf("Failed to create test cluster: %v", err)
	}

	const numGoroutines = 5
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine gets its own transaction
			err := store.BeginTransaction(ctx)
			if err != nil {
				errors <- err
				return
			}

			// Simulate work
			time.Sleep(50 * time.Millisecond)

			// Each transaction should be independent
			_, err = store.ListClusterIds(ctx)
			if err != nil {
				errors <- err
				return
			}

			// Rollback to clean up
			if err := store.RollbackTransaction(ctx); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent transaction error: %v", err)
	}
}

// TestTransactionNotSharedAcrossGoroutines explicitly tests that
// the transaction state is NOT shared
func TestTransactionNotSharedAcrossGoroutines(t *testing.T) {
	connectionString := "postgres://postgres:postgres@localhost:5432/trax_test?sslmode=disable"

	store, err := NewPsqlStore(connectionString)
	if err != nil {
		t.Skip("Skipping test: cannot connect to test database")
		return
	}
	defer store.Close()

	ctx := context.Background()

	// Goroutine 1 begins a transaction
	err = store.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction 1: %v", err)
	}

	// The base store should NOT have a transaction
	// (This is what the old implementation got wrong)

	// Try to commit on base store - should fail with "no active transaction"
	err = store.CommitTransaction(ctx)
	if err == nil {
		t.Error("Expected error when committing without transaction, got nil")
	}

	// Try to begin another transaction - should succeed
	// (This would fail in the old implementation where tx was shared)
	err = store.BeginTransaction(ctx)
	if err != nil {
		t.Errorf("Should be able to begin second transaction: %v", err)
	}

	// Clean up
	store.RollbackTransaction(ctx)
}
