ALTER TABLE order_services
    ADD COLUMN start_date TIMESTAMPTZ,
    ADD COLUMN end_date TIMESTAMPTZ NULL,
    ADD COLUMN observations TEXT NULL;

UPDATE order_services
SET start_date = created_at
WHERE start_date IS NULL;

ALTER TABLE order_services
    ALTER COLUMN start_date SET NOT NULL;
