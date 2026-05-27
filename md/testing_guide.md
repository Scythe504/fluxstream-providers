# Go Unit Testing Guide for fluxstream-providers

This guide explains how to write robust, isolated, and lightning-fast unit tests for the three primary layers of the `fluxstream-providers` microservice:
1. **Database Layer** (SQLite CRUD)
2. **HTTP Server Layer** (gorilla/mux Handlers)
3. **Verification Worker Layer** (Mocking HTTP feeds)

---

## 1. Testing the Database Layer

Instead of using Docker/Testcontainers (which slow down builds and have external dependencies), write database tests against a temporary local SQLite file.

### Example Database Test (`database_test.go`)
```go
package database

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// 1. Override the global package DSN to a temporary file
	dburl = "test_providers.db"
	
	// 2. Run the tests
	code := m.Run()
	
	// 3. Close the active database connection
	if dbInstance != nil && dbInstance.db != nil {
		_ = dbInstance.Close()
	}
	
	// 4. Remove the temporary file
	_ = os.Remove("test_providers.db")
	
	os.Exit(code)
}

func TestProviderCRUD(t *testing.T) {
	dbInstance = nil
	srv := New()
	ctx := context.Background()

	p := &Provider{
		ID:                  "test-uuid-1",
		ProviderName:        "TestProvider",
		ProviderURL:         "http://localhost:9091",
		VerificationPending: true,
		Version:             "1.0.0",
		VerifiedAt:          nil,
		ProviderType:        "anime",
		CreatedAt:           time.Now().Unix(),
	}

	// Verify Insert
	if err := srv.UpsertProvider(ctx, p); err != nil {
		t.Fatalf("expected no error inserting provider, got %v", err)
	}

	// Verify Retrieve
	retrieved, err := srv.GetProvider(ctx, p.ID)
	if err != nil || retrieved.ProviderName != "TestProvider" {
		t.Fatalf("failed to retrieve correct provider: %v", err)
	}
}
```

---

## 2. Testing HTTP Route Handlers

Use standard Go `"net/http/httptest"` to record requests and verify response status codes, headers, and bodies without booting up a live socket listener.

### Example Route Handler Test (`routes_test.go`)
```go
package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetProviderHandler(t *testing.T) {
	// 1. Set up your server with a mock/test database
	s := &Server{
		db: newMockDatabaseService(), // implement a mock satisfying database.Service
	}

	// 2. Create the request
	req, err := http.NewRequest("GET", "/api/providers/uuid-1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// 3. Create a response recorder
	rr := httptest.NewRecorder()
	
	// 4. Invoke the handler
	s.getProviderHandler(rr, req)

	// 5. Assertions
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var p database.Provider
	if err := json.NewDecoder(rr.Body).Decode(&p); err != nil {
		t.Errorf("failed to decode JSON response: %v", err)
	}
}
```

---

## 3. Testing the Verification Worker

To test `resolver.VerifyProviderURL` without reaching out to an actual live internet provider, spin up a local dummy server via `httptest.NewServer` that serves the mock type contract.

### Example Verification Test (`verifier_test.go`)
```go
package resolver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifyProviderURL(t *testing.T) {
	// 1. Set up a local test HTTP server to mimic a provider
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock the `/api/trending` path
		if r.URL.Path == "/api/trending" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			
			// Return a valid mock Media slice
			mockMedia := []Media{
				{
					ID:    "123",
					Type:  MediaTypeAnime,
					Title: "Mock Anime",
				},
			}
			_ = json.NewEncoder(w).Encode(mockMedia)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockProvider.Close()

	// 2. Call verifier on the mock server URL
	ctx := context.Background()
	version, err := VerifyProviderURL(ctx, mockProvider.URL)
	if err != nil {
		t.Fatalf("expected verification to pass, got error: %v", err)
	}

	if version != "1.0.0" {
		t.Errorf("expected version to be '1.0.0', got '%s'", version)
	}
}
```
