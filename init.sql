CREATE database "9eyes";
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
\c "9eyes";

BEGIN;

-- Table for measuring distance to cats over time.
CREATE TABLE distance (
  time      TIMESTAMPTZ       NOT NULL,
  location  TEXT              NOT NULL,
  cat       TEXT              NOT NULL,
  distance  DOUBLE PRECISION  NOT NULL
);
SELECT create_hypertable('distance', 'time');

-- Table for measuring scale weight over time.
CREATE TABLE scale (
  time      TIMESTAMPTZ       NOT NULL,
  location  TEXT              NOT NULL,
  weight_g  DOUBLE PRECISION  NOT NULL
);
SELECT create_hypertable('scale', 'time');

COMMIT;
