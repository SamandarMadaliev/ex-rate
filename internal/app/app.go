package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

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

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	app := App{
		config: config,
		db:     db,
		server: &http.Server{
			Addr:    config.Server.Host + ":" + config.Server.Port,
			Handler: mux,
		},
	}

	return &app, nil
}

func (a App) Run() error {
	return a.server.ListenAndServe()
}

func (a App) Stop() error {
	if err := a.server.Shutdown(context.Background()); err != nil {
		return err
	}

	return a.db.Close()
}
