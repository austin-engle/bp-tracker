-- File: internal/database/schema.sql

CREATE TABLE IF NOT EXISTS readings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    systolic INTEGER NOT NULL,
    diastolic INTEGER NOT NULL,
    pulse INTEGER NOT NULL,
    classification TEXT NOT NULL,
    CONSTRAINT valid_systolic CHECK (systolic BETWEEN 60 AND 250),
    CONSTRAINT valid_diastolic CHECK (diastolic BETWEEN 40 AND 150),
    CONSTRAINT valid_pulse CHECK (pulse BETWEEN 40 AND 200)
);

-- Index for faster querying of recent readings
CREATE INDEX IF NOT EXISTS idx_readings_timestamp ON readings(timestamp);
