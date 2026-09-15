package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	router "github.com/SamandarMadaliev/ex-rate/internal/http"
	"github.com/SamandarMadaliev/ex-rate/internal/repositories"
	"github.com/SamandarMadaliev/ex-rate/internal/services"
	"github.com/SamandarMadaliev/ex-rate/internal/worker"
	"github.com/SamandarMadaliev/ex-rate/pkg/config"
	"github.com/SamandarMadaliev/ex-rate/pkg/database/postgres"
)

type App struct {
	config  *config.Config
	db      *sql.DB
	server  *http.Server
	workers *worker.Pool
}

func NewApp(config *config.Config) (*App, error) {
	db, err := postgres.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	currencyRepo := repositories.NewCurrencyRepository(db)
	rateRepo := repositories.NewRateRepository(db)

	workers := worker.NewPool(config.Worker.Count, config.Worker.BufferSize)
	priceService := services.NewService(nil, config.ExRateAPI.URL, config.ExRateAPI.Token)

	routes := router.NewRouter(db, currencyRepo, rateRepo, workers, priceService)

	server := &http.Server{
		Addr:              config.Server.Host + ":" + config.Server.Port,
		Handler:           routes,
		ReadTimeout:       config.Server.ReadTimeout,
		WriteTimeout:      config.Server.WriteTimeout,
		IdleTimeout:       config.Server.IdleTimeout,
		ReadHeaderTimeout: config.Server.ReadHeaderTimeout,
	}

	app := App{
		config:  config,
		db:      db,
		server:  server,
		workers: workers,
	}

	return &app, nil
}

func (a App) SubmitJob(job worker.Job) error {
	return a.workers.Submit(job)
}

func (a App) Run() error {
	return a.server.ListenAndServe()
}

func (a App) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.Server.ShutdownTimeout)
	defer cancel()

	shutdownErr := a.server.Shutdown(ctx)
	a.workers.Stop()

	closeErr := a.db.Close()

	return errors.Join(shutdownErr, closeErr)
}
