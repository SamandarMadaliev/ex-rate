CREATE TABLE currencies (
    id SERIAL PRIMARY KEY,
    slug TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_currencies_slug ON currencies (slug);

INSERT INTO currencies (slug) VALUES ('UZS'), ('USD'), ('EUR');
