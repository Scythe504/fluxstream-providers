package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/scythe504/fluxstream-providers/internal/database"
	"github.com/scythe504/fluxstream-providers/internal/media"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	// Apply CORS middleware
	r.Use(s.corsMiddleware)

	r.HandleFunc("/", s.HelloWorldHandler)
	r.HandleFunc("/health", s.healthHandler)

	// Providers endpoints
	r.HandleFunc("/api/providers", s.listProvidersHandler).Methods("GET")
	r.HandleFunc("/api/providers", s.upsertProviderHandler).Methods("POST", "PUT")
	r.HandleFunc("/api/providers/{id}", s.getProviderHandler).Methods("GET")

	return r
}

// CORS middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS Headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Wildcard allows all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Credentials not allowed with wildcard origins

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

type UpsertProviderRequest struct {
	Name string `json:"provider_name"`
	URL  string `json:"provider_url"`
	Type string `json:"provider_type"`
}

func (s *Server) getProviderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing provider id", http.StatusBadRequest)
		return
	}

	p, err := s.db.GetProvider(r.Context(), id)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

func (s *Server) listProvidersHandler(w http.ResponseWriter, r *http.Request) {
	page, perPage, err := getPageParams(r, 1, 24)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	limit := perPage
	offset := (page - 1) * perPage

	list, err := s.db.ListProviders(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list providers: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (s *Server) upsertProviderHandler(w http.ResponseWriter, r *http.Request) {
	var req UpsertProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request JSON payload", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.URL == "" || req.Type == "" {
		http.Error(w, "provider_name, provider_url, and provider_type are required", http.StatusBadRequest)
		return
	}

	// Check if a provider with this name already exists
	existing, err := s.db.GetProviderByName(r.Context(), req.Name)
	var p *database.Provider
	if err == nil && existing != nil {
		// Existing provider: update URL, Type, and reset verification states
		p = existing
		p.ProviderURL = req.URL
		p.ProviderType = req.Type
		p.VerificationPending = true
		p.VerifiedAt = nil
	} else {
		// New provider: generate brand new UUID and set timestamps
		p = &database.Provider{
			ID:                  uuid.New().String(),
			ProviderName:        req.Name,
			ProviderURL:         req.URL,
			VerificationPending: true,
			Version:             "1.0.0",
			VerifiedAt:          nil,
			ProviderType:        req.Type,
			CreatedAt:           time.Now().Unix(),
		}
	}

	// Perform DB Upsert
	if err := s.db.UpsertProvider(r.Context(), p); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save provider: %v", err), http.StatusInternalServerError)
		return
	}

	// Launch the background verification worker asynchronously
	go func(prov database.Provider) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		log.Printf("[Verification Worker] Starting verification for provider: %s (%s)...", prov.ProviderName, prov.ProviderURL)
		version, err := media.VerifyProviderURL(ctx, prov.ProviderURL)
		if err != nil {
			log.Printf("[Verification Worker] FAILED for %s: %v. Provider remains unverified.", prov.ProviderName, err)
			return
		}

		// Success! Update DB verification status
		now := time.Now().Unix()
		prov.VerificationPending = false
		prov.VerifiedAt = &now
		prov.Version = version

		if err := s.db.UpsertProvider(context.Background(), &prov); err != nil {
			log.Printf("[Verification Worker] DB Update failed for %s: %v", prov.ProviderName, err)
		} else {
			log.Printf("[Verification Worker] SUCCESS! Provider %s verified and activated successfully (version: %s)", prov.ProviderName, version)
		}
	}(*p)

	// Return registered/upserted provider details (Accepted state)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(p)
}

func getPageParams(r *http.Request, defaultPage, defaultPerPage int) (int, int, error) {
	page, err := getPositiveIntQuery(r, "page", defaultPage)
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid page query, value must be a positive number")
	}

	perPage, err := getPositiveIntQuery(r, "perPage", defaultPerPage)
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid perPage query, value must be a positive number")
	}

	return page, perPage, nil
}

func getPositiveIntQuery(r *http.Request, key string, defaultValue int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if value < 1 {
		return 0, fmt.Errorf("%s must be positive", key)
	}

	return value, nil
}
