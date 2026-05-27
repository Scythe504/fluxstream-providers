package database

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Override package-level dburl to a local temporary SQLite file for tests
	dburl = "test_providers.db"
	
	// Run all tests
	code := m.Run()
	
	// Close any active instance before removing the file
	if dbInstance != nil && dbInstance.db != nil {
		_ = dbInstance.Close()
	}
	
	// Clean up the temporary database file
	_ = os.Remove("test_providers.db")
	
	os.Exit(code)
}

func TestNew(t *testing.T) {
	// Reset the global database instance for a clean test run
	dbInstance = nil
	
	srv := New()
	if srv == nil {
		t.Fatal("New() returned nil")
	}
}

func TestHealth(t *testing.T) {
	dbInstance = nil
	srv := New()

	stats := srv.Health()

	if stats["status"] != "up" {
		t.Fatalf("expected status to be up, got %s", stats["status"])
	}

	if _, ok := stats["error"]; ok {
		t.Fatalf("expected error not to be present")
	}

	if stats["message"] != "It's healthy" {
		t.Fatalf("expected message to be 'It's healthy', got %s", stats["message"])
	}
}

func TestClose(t *testing.T) {
	dbInstance = nil
	srv := New()

	if srv.Close() != nil {
		t.Fatalf("expected Close() to return nil")
	}
}

func TestProviderCRUD(t *testing.T) {
	dbInstance = nil
	srv := New()
	ctx := context.Background()

	// 1. Create a dummy provider
	p := &Provider{
		ID:                  "test-uuid-1",
		ProviderName:        "TestAni",
		ProviderURL:         "http://localhost:9091",
		VerificationPending: true,
		Version:             "1.0.0",
		VerifiedAt:          nil,
		ProviderType:        "anime",
		CreatedAt:           time.Now().Unix(),
	}

	// Test Upsert
	err := srv.UpsertProvider(ctx, p)
	if err != nil {
		t.Fatalf("expected no error upserting provider, got %v", err)
	}

	// Test Get
	retrieved, err := srv.GetProvider(ctx, "test-uuid-1")
	if err != nil {
		t.Fatalf("expected no error getting provider, got %v", err)
	}
	if retrieved.ProviderName != "TestAni" {
		t.Errorf("expected provider_name to be 'TestAni', got %s", retrieved.ProviderName)
	}

	// Test GetByName
	retrievedByName, err := srv.GetProviderByName(ctx, "TestAni")
	if err != nil {
		t.Fatalf("expected no error getting provider by name, got %v", err)
	}
	if retrievedByName.ID != "test-uuid-1" {
		t.Errorf("expected provider ID to be 'test-uuid-1', got %s", retrievedByName.ID)
	}

	// Test List
	list, err := srv.ListProviders(ctx, 10, 0)
	if err != nil {
		t.Fatalf("expected no error listing providers, got %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected list length to be 1, got %d", len(list))
	}

	// Test Upsert (update url and type)
	p.ProviderURL = "http://localhost:9092"
	p.ProviderType = "series"
	err = srv.UpsertProvider(ctx, p)
	if err != nil {
		t.Fatalf("expected no error updating provider, got %v", err)
	}

	retrievedUpdated, err := srv.GetProvider(ctx, "test-uuid-1")
	if err != nil {
		t.Fatalf("expected no error getting updated provider, got %v", err)
	}
	if retrievedUpdated.ProviderURL != "http://localhost:9092" {
		t.Errorf("expected updated URL to be 'http://localhost:9092', got %s", retrievedUpdated.ProviderURL)
	}
	if retrievedUpdated.ProviderType != "series" {
		t.Errorf("expected updated Type to be 'series', got %s", retrievedUpdated.ProviderType)
	}
}
