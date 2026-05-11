ALTER TABLE order_services
    DROP COLUMN IF EXISTS observations,
    DROP COLUMN IF EXISTS end_date,
    DROP COLUMN IF EXISTS start_date;
