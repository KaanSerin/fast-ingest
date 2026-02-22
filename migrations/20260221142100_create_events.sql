CREATE TABLE IF NOT EXISTS events (
  id           BIGSERIAL PRIMARY KEY,
  event_name   TEXT        NOT NULL,
  channel      TEXT        NOT NULL,
  campaign_id  TEXT        NULL,
  user_id      TEXT        NOT NULL,
  ts           TIMESTAMPTZ NOT NULL,

  tags         JSONB       NOT NULL DEFAULT '{}'::jsonb,
  metadata     JSONB       NOT NULL DEFAULT '{}'::jsonb,

  ingested_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create indexes to optimize queries based on common access patterns.
CREATE INDEX IF NOT EXISTS ix_events_user_ts
  ON events (user_id, ts DESC);

CREATE INDEX IF NOT EXISTS ix_events_campaign_ts
  ON events (campaign_id, ts DESC)
  WHERE campaign_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_events_name_ts
  ON events (event_name, ts DESC);

CREATE INDEX IF NOT EXISTS ix_events_channel_ts
  ON events (channel, ts DESC);

CREATE INDEX IF NOT EXISTS ix_events_eventname_ts_user
  ON events (event_name, ts DESC, user_id);