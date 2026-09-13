package handlers

import (
	"database/sql"
	"net/http"

	"github.com/SamandarMadaliev/ex-rate/internal/http/helpers"
)

func HealthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			helpers.WriteError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
