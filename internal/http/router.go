package http

import (
	"database/sql"
	"net/http"

	"github.com/SamandarMadaliev/ex-rate/internal/http/handlers"
	"github.com/SamandarMadaliev/ex-rate/internal/repositories"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(db *sql.DB, currencyRepo *repositories.CurrencyRepository, rateRepo *repositories.RateRepository) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.SetHeader("Content-Type", "application/json"))

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handlers.HealthHandler(db))
		r.Get("/currencies", handlers.CurrenciesHandler(currencyRepo))
		r.Post("/rates", handlers.CreateRateHandler(rateRepo, currencyRepo))
		r.Get("/rates/latest", handlers.LatestRateHandler(rateRepo))
		r.Get("/rates/{id}", handlers.GetRateHandler(rateRepo))
	})

	return router
}
