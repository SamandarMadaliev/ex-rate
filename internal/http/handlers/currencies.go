package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/SamandarMadaliev/ex-rate/internal/http/helpers"
	"github.com/SamandarMadaliev/ex-rate/internal/repositories"
)

func CurrenciesHandler(repo *repositories.CurrencyRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currencies, err := repo.List(r.Context())
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "failed to list currencies")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(currencies)
	}
}
