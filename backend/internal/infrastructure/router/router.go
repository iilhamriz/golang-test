package router

import (
	"github.com/calyx/smart-inventory/internal/adapter/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func New(
	siHandler *handler.StockInHandler,
	soHandler *handler.StockOutHandler,
	invHandler *handler.InventoryHandler,
	custHandler *handler.CustomerHandler,
	reportHandler *handler.ReportHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		// Stock In
		r.Route("/stock-in", func(r chi.Router) {
			r.Post("/", siHandler.Create)
			r.Get("/", siHandler.List)
			r.Get("/{id}", siHandler.GetByID)
			r.Patch("/{id}/status", siHandler.UpdateStatus)
			r.Post("/{id}/cancel", siHandler.Cancel)
		})

		// Stock Out
		r.Route("/stock-out", func(r chi.Router) {
			r.Post("/", soHandler.Create)
			r.Get("/", soHandler.List)
			r.Get("/{id}", soHandler.GetByID)
			r.Patch("/{id}/status", soHandler.UpdateStatus)
			r.Post("/{id}/cancel", soHandler.Cancel)
		})

		// Inventory / Items
		r.Route("/items", func(r chi.Router) {
			r.Post("/", invHandler.Create)
			r.Get("/", invHandler.List)
			r.Get("/{id}", invHandler.GetByID)
			r.Put("/{id}", invHandler.Update)
			r.Post("/{id}/adjust", invHandler.Adjust)
		})

		// Customers
		r.Route("/customers", func(r chi.Router) {
			r.Post("/", custHandler.Create)
			r.Get("/", custHandler.List)
			r.Get("/{id}", custHandler.GetByID)
		})

		// Reports
		r.Route("/reports", func(r chi.Router) {
			r.Get("/transactions", reportHandler.ListDoneTransactions)
			r.Get("/transactions/{type}/{id}", reportHandler.GetTransactionDetail)
		})
	})

	return r
}
