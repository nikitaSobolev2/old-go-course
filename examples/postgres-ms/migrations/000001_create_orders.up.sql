CREATE TABLE orders (
  id UUID PRIMARY KEY,
  customer_id TEXT NOT NULL,
  status TEXT NOT NULL,
  items JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);
