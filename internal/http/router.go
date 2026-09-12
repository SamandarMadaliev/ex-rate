package http

import (
	"database/sql"
	"net/http"

	"github.com/SamandarMadaliev/ex-rate/internal/http/handlers"
	"github.com/go-chi/chi/v5"
)

func NewRouter(db *sql.DB) http.Handler {
	router := chi.NewRouter()

	router.Get("/health", handlers.HealthHandler(db))

	return router
}
