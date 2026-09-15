package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/SamandarMadaliev/ex-rate/internal/http/helpers"
	// Referenced only in swag doc comments below.
	_ "github.com/SamandarMadaliev/ex-rate/internal/http/schemas"
)

// HealthHandler godoc
// @Summary      Health check
// @Description  Pings the database to verify the service is ready to serve traffic
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string  "OK"
// @Failure      503  {object}  schemas.ErrorResponse
// @Router       /health [get]
func HealthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			log.Printf("health: database unavailable: %v", err)
			helpers.WriteError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
