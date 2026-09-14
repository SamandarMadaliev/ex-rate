package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/SamandarMadaliev/ex-rate/internal/http/helpers"
	// Referenced only in swag doc comments below.
	_ "github.com/SamandarMadaliev/ex-rate/internal/http/schemas"
	_ "github.com/SamandarMadaliev/ex-rate/internal/models"
	"github.com/SamandarMadaliev/ex-rate/internal/repositories"
)

// CurrenciesHandler godoc
// @Summary      List currencies
// @Description  Returns all currencies known to the service
// @Tags         currencies
// @Produce      json
// @Success      200  {array}   models.Currency
// @Failure      500  {object}  schemas.ErrorResponse
// @Router       /currencies [get]
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
