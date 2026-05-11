ALTER TABLE orders
    ADD COLUMN deleted_at TIMESTAMPTZ NULL;

CREATE INDEX idx_orders_deleted_at ON orders (deleted_at);
