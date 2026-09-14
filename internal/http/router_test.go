package http_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	router "github.com/SamandarMadaliev/ex-rate/internal/http"
	"github.com/SamandarMadaliev/ex-rate/internal/http/schemas"
	"github.com/SamandarMadaliev/ex-rate/internal/models"
	"github.com/SamandarMadaliev/ex-rate/internal/repositories"
	"github.com/SamandarMadaliev/ex-rate/internal/services"
	"github.com/SamandarMadaliev/ex-rate/internal/worker"
)

// newTestRouter wires the real router against a sqlmock-backed *sql.DB. The
// worker pool is stopped immediately after creation so job submission is
// deterministic (always ErrPoolStopped, logged and ignored by the handler)
// instead of racing a background goroutine against the mock's expectations.
func newTestRouter(t *testing.T) (http.Handler, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	currencyRepo := repositories.NewCurrencyRepository(db)
	rateRepo := repositories.NewRateRepository(db)

	workers := worker.NewPool(1, 1)
	workers.Stop()

	priceService := services.NewService(nil, "", "")

	return router.NewRouter(db, currencyRepo, rateRepo, workers, priceService), mock
}

func decodeJSON(t *testing.T, body *bytes.Buffer, v any) {
	t.Helper()
	if err := json.NewDecoder(body).Decode(v); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

func assertExpectationsMet(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestHealthHandler(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		mock.ExpectPing()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if rec.Body.String() != "OK" {
			t.Fatalf("expected body %q, got %q", "OK", rec.Body.String())
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("database unavailable", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		mock.ExpectPing().WillReturnError(errors.New("connection refused"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})
}

func TestCurrenciesHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		rows := sqlmock.NewRows([]string{"id", "slug", "is_active"}).
			AddRow(int64(1), "USD", true).
			AddRow(int64(2), "EUR", true)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, slug, is_active FROM currencies ORDER BY id")).
			WillReturnRows(rows)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/currencies", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var currencies []models.Currency
		decodeJSON(t, rec.Body, &currencies)
		if len(currencies) != 2 {
			t.Fatalf("expected 2 currencies, got %d", len(currencies))
		}
		if currencies[0].Slug != "USD" || currencies[1].Slug != "EUR" {
			t.Fatalf("unexpected currencies: %+v", currencies)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, slug, is_active FROM currencies ORDER BY id")).
			WillReturnError(errors.New("db down"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/currencies", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})
}

func TestCreateRateHandler(t *testing.T) {
	body := func(base, quote int64) *bytes.Buffer {
		b, _ := json.Marshal(schemas.CreateRateRequest{BaseCurrencyID: base, QuoteCurrencyID: quote})
		return bytes.NewBuffer(b)
	}

	t.Run("invalid JSON body", func(t *testing.T) {
		handler, mock := newTestRouter(t)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rates", bytes.NewBufferString("{not-json"))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("validation failure", func(t *testing.T) {
		handler, mock := newTestRouter(t)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rates", body(1, 1))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("currency validation query fails", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM currencies WHERE id IN")).
			WillReturnError(errors.New("db down"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rates", body(1, 2))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("base currency invalid", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(2))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM currencies WHERE id IN")).
			WillReturnRows(rows)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rates", body(1, 2))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("quote currency invalid", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(1))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM currencies WHERE id IN")).
			WillReturnRows(rows)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rates", body(1, 2))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("rate creation fails", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		validRows := sqlmock.NewRows([]string{"id"}).AddRow(int64(1)).AddRow(int64(2))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM currencies WHERE id IN")).
			WillReturnRows(validRows)
		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO rates")).
			WillReturnError(errors.New("db down"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rates", body(1, 2))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("success", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		validRows := sqlmock.NewRows([]string{"id"}).AddRow(int64(1)).AddRow(int64(2))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM currencies WHERE id IN")).
			WillReturnRows(validRows)

		createdRows := sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("11111111-1111-1111-1111-111111111111", time.Now())
		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO rates")).
			WillReturnRows(createdRows)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rates", body(1, 2))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", rec.Code)
		}

		var resp schemas.CreateRateResponse
		decodeJSON(t, rec.Body, &resp)
		if resp.ID != "11111111-1111-1111-1111-111111111111" {
			t.Fatalf("unexpected rate id: %q", resp.ID)
		}
		assertExpectationsMet(t, mock)
	})
}

func TestGetRateHandler(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		handler, mock := newTestRouter(t)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rates/not-a-uuid", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		id := "11111111-1111-1111-1111-111111111111"
		mock.ExpectQuery(regexp.QuoteMeta("FROM rates")).
			WillReturnError(sql.ErrNoRows)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rates/"+id, nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("database error", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		id := "11111111-1111-1111-1111-111111111111"
		mock.ExpectQuery(regexp.QuoteMeta("FROM rates")).
			WillReturnError(errors.New("db down"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rates/"+id, nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("success", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		id := "11111111-1111-1111-1111-111111111111"
		price := 1.23
		rows := sqlmock.NewRows([]string{
			"id", "base_currency_id", "quote_currency_id", "price", "status",
			"price_timestamp", "updated_at", "created_at",
		}).AddRow(id, int64(1), int64(2), price, string(models.RateStatusCompleted), time.Now(), time.Now(), time.Now())
		mock.ExpectQuery(regexp.QuoteMeta("FROM rates")).
			WillReturnRows(rows)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rates/"+id, nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var rate models.Rate
		decodeJSON(t, rec.Body, &rate)
		if rate.ID != id || rate.Price == nil || *rate.Price != price {
			t.Fatalf("unexpected rate: %+v", rate)
		}
		assertExpectationsMet(t, mock)
	})
}

func TestLatestRateHandler(t *testing.T) {
	t.Run("missing query params", func(t *testing.T) {
		handler, mock := newTestRouter(t)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rates/latest", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		mock.ExpectQuery(regexp.QuoteMeta("FROM rates")).
			WillReturnError(sql.ErrNoRows)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rates/latest?base=1&quote=2", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
		assertExpectationsMet(t, mock)
	})

	t.Run("success", func(t *testing.T) {
		handler, mock := newTestRouter(t)
		id := "11111111-1111-1111-1111-111111111111"
		rows := sqlmock.NewRows([]string{
			"id", "base_currency_id", "quote_currency_id", "price", "status",
			"price_timestamp", "updated_at", "created_at",
		}).AddRow(id, int64(1), int64(2), 1.5, string(models.RateStatusCompleted), time.Now(), time.Now(), time.Now())
		mock.ExpectQuery(regexp.QuoteMeta("FROM rates")).
			WillReturnRows(rows)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rates/latest?base=1&quote=2", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var rate models.Rate
		decodeJSON(t, rec.Body, &rate)
		if rate.ID != id {
			t.Fatalf("unexpected rate: %+v", rate)
		}
		assertExpectationsMet(t, mock)
	})
}
