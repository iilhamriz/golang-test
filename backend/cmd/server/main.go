package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	adapterRepo "github.com/calyx/smart-inventory/internal/adapter/repository"
	"github.com/calyx/smart-inventory/internal/adapter/handler"
	"github.com/calyx/smart-inventory/internal/infrastructure/config"
	"github.com/calyx/smart-inventory/internal/infrastructure/database"
	"github.com/calyx/smart-inventory/internal/infrastructure/router"
	"github.com/calyx/smart-inventory/internal/usecase"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	ctx := context.Background()

	// Database
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to database")

	// Run migrations
	exe, _ := os.Executable()
	migrationsDir := filepath.Join(filepath.Dir(exe), "..", "..", "migrations")
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "migrations"
	}
	if err := database.RunMigrations(ctx, pool, migrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Repositories
	itemRepo := adapterRepo.NewPostgresItemRepo(pool)
	customerRepo := adapterRepo.NewPostgresCustomerRepo(pool)
	stockInRepo := adapterRepo.NewPostgresStockInRepo(pool)
	stockOutRepo := adapterRepo.NewPostgresStockOutRepo(pool)
	logRepo := adapterRepo.NewPostgresTransactionLogRepo(pool)
	adjRepo := adapterRepo.NewPostgresStockAdjustmentRepo(pool)

	// Use Cases
	stockInUC := usecase.NewStockInUseCase(pool, stockInRepo, itemRepo, logRepo)
	stockOutUC := usecase.NewStockOutUseCase(pool, stockOutRepo, itemRepo, logRepo)
	inventoryUC := usecase.NewInventoryUseCase(pool, itemRepo, adjRepo)
	reportUC := usecase.NewReportUseCase(pool, stockInRepo, stockOutRepo, logRepo)

	// Handlers
	siHandler := handler.NewStockInHandler(stockInUC)
	soHandler := handler.NewStockOutHandler(stockOutUC)
	invHandler := handler.NewInventoryHandler(inventoryUC)
	custHandler := handler.NewCustomerHandler(customerRepo)
	reportHandler := handler.NewReportHandler(reportUC)

	// Router
	r := router.New(siHandler, soHandler, invHandler, custHandler, reportHandler)

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
	log.Println("Server stopped")
}
