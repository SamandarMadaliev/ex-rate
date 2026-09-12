package postgres

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/SamandarMadaliev/ex-rate/pkg/config"
)

func New(cfg *config.Config) (*sql.DB, error) {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.DB.User, cfg.DB.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.DB.Host, cfg.DB.Port),
		Path:   "/" + cfg.DB.Name,
		RawQuery: url.Values{
			"sslmode": {cfg.DB.SSLMode},
		}.Encode(),
	}

	db, err := sql.Open("pgx", dsn.String())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
