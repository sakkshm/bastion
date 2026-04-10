package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sakkshm/bastion/internal/api"
	"github.com/sakkshm/bastion/internal/config"
	"github.com/sakkshm/bastion/internal/engine"
	"github.com/sakkshm/bastion/internal/logger"
)

func main() {

	// Flags for config file
	var configPath string
	flag.StringVar(&configPath, "config", "config/config.toml", "Config TOML file")
	flag.Parse()

	//  Load Config
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	//  Initialize Logger
	log, err := logger.New(
		cfg.Logging.Level,
		cfg.Logging.Format,
		cfg.Logging.File,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log.Info(
		"Starting Bastion server",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
	)

	// Inititalise runtime
	log.Info("Initializing engine")
	eng, err := engine.NewEngine(cfg, log)
	if err != nil {
		log.Error("Failed to initialize runtime", "error", err)
		os.Exit(1)
	}

	// Route Handlers
	routeHandler := api.NewHandler(eng)

	//  Router
	r := chi.NewRouter()

	// Attach Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Bastion server running!"))
	})

	r.Post(api.CreateSessionEndpoint, routeHandler.CreateNewSession)

	r.Route(api.SessionBaseEndpoint, func(r chi.Router) {
		r.Use(routeHandler.SessionMiddleware)

		// Session Endpoints
		r.Post(api.StartSessionEndpoint, routeHandler.StartSessionHandler)
		r.Post(api.StopSessionEndpoint, routeHandler.StopSessionHandler)
		r.Delete(api.DeleteSessionEndpoint, routeHandler.DeleteSessionHandler)
		r.Get(api.GetSessionStatusEndpoint, routeHandler.GetSessionStatusHandler)

		// Job endpoints
		r.Post(api.JobExecuteEndpoint, routeHandler.JobExecuteHandler)
		r.Get(api.GetJobStatusEndpoint, routeHandler.GetJobStatusHandler)

		// WS terminal endpoint
		r.Get(api.TerminalEndpoint, routeHandler.TerminalHandler)

		// Filesystem endpoint
		r.Post(api.UploadEndpoint, routeHandler.UploadHandler)
		r.Get(api.DownloadEndpoint, routeHandler.DownloadHandler)
		r.Delete(api.DeleteEndpoint, routeHandler.DeleteHandler)
	})

	port := fmt.Sprintf(":%d", cfg.Server.Port)

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	//  Start Server
	go func() {
		log.Info("Server is now listening", "address", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server crashed unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	//  Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Info("Shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Graceful shutdown failed", "error", err)
	} else {
		log.Info("Server stopped gracefully")
	}

	if err := eng.Close(); err != nil {
		log.Error("Failed to close client resources", "error", err)
	} else {
		log.Info("Client resources closed")
	}
}
