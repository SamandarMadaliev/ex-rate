package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	router "github.com/SamandarMadaliev/ex-rate/internal/http"
	"github.com/SamandarMadaliev/ex-rate/internal/repositories"
	"github.com/SamandarMadaliev/ex-rate/pkg/config"
	"github.com/SamandarMadaliev/ex-rate/pkg/database/postgres"
)

type App struct {
	config *config.Config
	db     *sql.DB
	server *http.Server
}

func NewApp(config *config.Config) (*App, error) {
	db, err := postgres.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	currencyRepo := repositories.NewCurrencyRepository(db)
	rateRepo := repositories.NewRateRepository(db)

	routes := router.NewRouter(db, currencyRepo, rateRepo)

	server := &http.Server{
		Addr:              config.Server.Host + ":" + config.Server.Port,
		Handler:           routes,
		ReadTimeout:       config.Server.ReadTimeout,
		WriteTimeout:      config.Server.WriteTimeout,
		IdleTimeout:       config.Server.IdleTimeout,
		ReadHeaderTimeout: config.Server.ReadHeaderTimeout,
	}

	app := App{
		config: config,
		db:     db,
		server: server,
	}

	return &app, nil
}

func (a App) Run() error {
	return a.server.ListenAndServe()
}

func (a App) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.Server.ShutdownTimeout)
	defer cancel()

	shutdownErr := a.server.Shutdown(ctx)
	closeErr := a.db.Close()

	return errors.Join(shutdownErr, closeErr)
}
