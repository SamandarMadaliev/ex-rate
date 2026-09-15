CREATE TABLE IF NOT EXISTS rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency_id INTEGER NOT NULL REFERENCES currencies(id),
    quote_currency_id INTEGER NOT NULL REFERENCES currencies(id),
    price NUMERIC(18,8),
    status TEXT NOT NULL DEFAULT 'pending',
    price_timestamp TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)