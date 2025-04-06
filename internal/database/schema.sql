-- File: internal/database/schema.sql (Modified for PostgreSQL)

CREATE TABLE IF NOT EXISTS readings (
    id SERIAL PRIMARY KEY, -- Changed from AUTOINCREMENT
    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Changed type and default
    systolic INTEGER NOT NULL,
    diastolic INTEGER NOT NULL,
    pulse INTEGER NOT NULL,
    classification VARCHAR NOT NULL, -- Changed from TEXT (VARCHAR is fine too)
    CONSTRAINT valid_systolic CHECK (systolic BETWEEN 60 AND 250),
    CONSTRAINT valid_diastolic CHECK (diastolic BETWEEN 40 AND 150),
    CONSTRAINT valid_pulse CHECK (pulse BETWEEN 40 AND 200)
);

-- Index for faster querying of recent readings
-- Syntax is the same for PostgreSQL
CREATE INDEX IF NOT EXISTS idx_readings_timestamp ON readings(timestamp);
