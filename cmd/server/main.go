package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/benjamin-rood/protogo-values-validation-demo/internal/server"
	v1 "github.com/benjamin-rood/protogo-values-validation-demo/gen/api/validation/v1"
	
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const (
	defaultPort = "8080"
	defaultGRPCPort = "9090"
)

func main() {
	// Get ports from environment or use defaults
	port := getEnvOrDefault("PORT", defaultPort)
	grpcPort := getEnvOrDefault("GRPC_PORT", defaultGRPCPort)

	// Create validation server
	validationServer := server.NewValidationServer()

	// Setup gRPC server
	grpcServer := grpc.NewServer()
	v1.RegisterValidationServiceServer(grpcServer, validationServer)
	
	// Add health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("validation.v1.ValidationService", grpc_health_v1.HealthCheckResponse_SERVING)
	
	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Start gRPC server
	grpcListener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	go func() {
		log.Printf("Starting gRPC server on port %s", grpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Setup HTTP health check endpoint
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/ready", readinessHandler(validationServer))

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: http.DefaultServeMux,
	}

	go func() {
		log.Printf("Starting HTTP server on port %s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Graceful stop gRPC server
	grpcServer.GracefulStop()

	log.Println("Servers stopped")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	
	response := `{
		"status": "healthy",
		"timestamp": "%s",
		"service": "protogo-values-validation-demo",
		"version": "1.0.0"
	}`
	
	fmt.Fprintf(w, response, time.Now().UTC().Format(time.RFC3339))
}

func readinessHandler(validationServer *server.ValidationServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Perform readiness checks
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Test the validation service
		req := &v1.ValidateTypesRequest{
			TestScenarios:  []string{"basic"},
			DeepValidation: false,
		}
		
		_, err := validationServer.ValidateTypes(ctx, req)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status": "not ready", "error": "%s"}`, err.Error())
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		
		response := `{
			"status": "ready",
			"timestamp": "%s",
			"service": "protogo-values-validation-demo"
		}`
		
		fmt.Fprintf(w, response, time.Now().UTC().Format(time.RFC3339))
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}