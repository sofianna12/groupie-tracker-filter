package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"groupie_tracker/backend/api"
	"groupie_tracker/backend/events"
	"groupie_tracker/backend/handlers"
	"groupie_tracker/backend/services"
	"groupie_tracker/db"
)

func initStore() db.Store {
	if url := os.Getenv("DB_URL"); url != "" {
		pg, err := db.NewPostgresStore(url)
		if err != nil {
			log.Fatalf("Cannot connect to postgres: %v", err)
		}
		log.Println("Using PostgreSQL store")
		return pg
	}
	log.Println("DB_URL not set — using in-memory store")
	return db.NewArtistStore()
}

func main() {
	// -----------------------------
	// Initialize Store
	// -----------------------------
	store := initStore()

	// -----------------------------
	// Preload External API Data (block server start so data is available immediately)
	// This is required for audits that expect the API to be populated at startup.
	// -----------------------------
	var dataLoaded = false

	// -----------------------------
	// Event System (Async Search)
	// -----------------------------
	searchChan := make(chan events.SearchEvent)
	go events.StartSearchWorker(store, searchChan)

	// -----------------------------
	// Services & Handlers
	// -----------------------------
	artistService := services.NewArtistService(store)
	artistHandler := handlers.NewArtistHandler(artistService, searchChan)

	// Perform synchronous preload. Fail fast if it can't load — audits expect data present.
	if err := api.FetchAndLoad(store); err != nil {
		log.Fatalf("Failed to preload external API: %v", err)
	}
	log.Println("External API data loaded successfully")
	dataLoaded = true

	mux := http.NewServeMux()

	// Register API routes
	artistHandler.RegisterRoutes(mux)

	// expose /api/loaded (will be true since preload completed)
	mux.HandleFunc("/api/loaded", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"loaded": dataLoaded})
	})

	// Health check endpoint (audit friendly)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Serve Angular static files
	spaDir := "./frontend/dist/frontend/browser"
	if _, err := os.Stat(spaDir); err == nil {
		fs := http.FileServer(http.Dir(spaDir))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// If path starts with /api, it's already handled above
			if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
				http.Error(w, "API endpoint not found", http.StatusNotFound)
				return
			}
			// Try to serve the file
			path := filepath.Join(spaDir, r.URL.Path)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				// File doesn't exist, serve index.html for Angular routing
				http.ServeFile(w, r, filepath.Join(spaDir, "index.html"))
				return
			}
			fs.ServeHTTP(w, r)
		})
		log.Println("Serving Angular frontend from", spaDir)
	} else {
		// Fallback if frontend not built
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
				http.Error(w, "API endpoint not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Frontend not built. Run: cd frontend && npm run build", http.StatusNotFound)
		})
		log.Println("Frontend not found. Build it with: cd frontend && npm run build")
	}

	// -----------------------------
	// HTTP Server Configuration
	// -----------------------------
	server := &http.Server{
		Addr:              ":8080",
		Handler:           handlers.RecoverMiddleware(mux),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// -----------------------------
	// Start Server
	// -----------------------------
	go func() {
		log.Println("Server running on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// -----------------------------
	// Graceful Shutdown
	// -----------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("⚠ Forced shutdown: %v\n", err)
	} else {
		log.Println("Server exited properly")
	}
}
