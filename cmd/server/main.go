package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/bartosz/homeboard/internal/config"
	"github.com/bartosz/homeboard/internal/handlers"
	"github.com/bartosz/homeboard/internal/widgets"
)

const (
	defaultConfigPath = "config.json"
	defaultPythonPath = "python3"
)

func main() {
	// Parse command line flags
	var (
		configPath = flag.String("config", defaultConfigPath, "Path to configuration file")
		pythonPath = flag.String("python", defaultPythonPath, "Path to Python interpreter")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Set up logging
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}

	log.Printf("Starting E-Paper Dashboard Server...")
	log.Printf("Config path: %s", *configPath)
	log.Printf("Python path: %s", *pythonPath)

	// Load initial configuration to validate and get server settings
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("Server will run on port %d", cfg.ServerPort)
	log.Printf("Dashboard refresh interval: %d minutes", cfg.RefreshInterval)
	log.Printf("Configured widgets: %d (%d enabled)", len(cfg.Widgets), len(cfg.GetEnabledWidgets()))

	// Create widget executor
	executor := widgets.NewExecutor(*pythonPath, 30*time.Second)

	// Validate all enabled widgets
	log.Printf("Validating widgets...")
	enabledWidgets := cfg.GetEnabledWidgets()
	for _, widget := range enabledWidgets {
		if err := executor.ValidateWidget(widget); err != nil {
			log.Printf("Warning: Widget '%s' validation failed: %v", widget.Name, err)
		} else {
			log.Printf("Widget '%s' validated successfully", widget.Name)
		}
	}

	// Create handlers
	dashboardHandler, err := handlers.NewDashboardHandler(*configPath, executor)
	if err != nil {
		log.Fatalf("Failed to create dashboard handler: %v", err)
	}

	adminHandler := handlers.NewAdminHandler(*configPath)

	// Set up HTTP routes
	router := mux.NewRouter()
	
	// Dashboard route
	router.Handle("/", dashboardHandler).Methods("GET")
	
	// Admin routes
	router.Handle("/admin", adminHandler).Methods("GET", "POST")
	router.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	})

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	}).Methods("GET")

	// API endpoint for configuration info
	router.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		currentCfg, err := config.LoadConfig(*configPath)
		if err != nil {
			http.Error(w, "Failed to load config", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		
		// Return safe config info (no sensitive parameters)
		safeConfig := map[string]interface{}{
			"title":            currentCfg.Title,
			"refresh_interval": currentCfg.RefreshInterval,
			"widget_count":     len(currentCfg.Widgets),
			"enabled_widgets":  len(currentCfg.GetEnabledWidgets()),
			"theme":           currentCfg.Theme,
		}
		
		if err := json.NewEncoder(w).Encode(safeConfig); err != nil {
			log.Printf("Error encoding config response: %v", err)
		}
	}).Methods("GET")

	// Set up HTTP server
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second, // Allow time for widget execution
		IdleTimeout:  120 * time.Second,
	}

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Printf("Shutdown signal received, stopping server...")
		
		// Give the server 30 seconds to finish any ongoing requests
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}()

	// Start the server
	log.Printf("Server starting on %s", cfg.GetServerAddress())
	log.Printf("Dashboard available at: http://localhost:%d/", cfg.ServerPort)
	log.Printf("Admin panel available at: http://localhost:%d/admin", cfg.ServerPort)
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Printf("Server stopped gracefully")
}