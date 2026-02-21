ALTER TABLE events ADD COLUMN IF NOT EXISTS dedupe_key TEXT NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_events_dedupe_key
  ON events (dedupe_key);