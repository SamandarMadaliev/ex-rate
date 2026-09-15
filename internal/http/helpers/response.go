package helpers

import (
	"encoding/json"
	"net/http"

	"github.com/SamandarMadaliev/ex-rate/internal/http/schemas"
)

func WriteError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(schemas.ErrorResponse{Error: message})
}
