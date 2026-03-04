-- +goose Up
-- +goose StatementBegin

-- Background: stop dates were originally stored as midnight UTC (T00:00:00Z)
-- because the frontend converted a bare YYYY-MM-DD input with:
--
--   `${val}T00:00:00Z`
--
-- Midnight UTC falls on the *previous* calendar day in every US timezone
-- (e.g. 00:00 UTC = 19:00 EST the evening before). This caused dates to
-- display one day earlier than entered in the UI.
--
-- The frontend was updated to store T17:00:00Z (noon EST = UTC-5) instead,
-- which keeps the instant safely within the entered calendar day from Hawaii
-- (09:00 HST) through Maine (12:00 EST).
--
-- This migration repairs existing rows that were stored with the old midnight
-- convention. Only rows whose time component is exactly 00:00:00 UTC are
-- touched — rows already stored at other times (e.g. real HH:MM data entered
-- manually) are left untouched.

UPDATE stops
SET arrived_at = arrived_at + INTERVAL '17 hours'
WHERE DATE_TRUNC('day', arrived_at AT TIME ZONE 'UTC') = arrived_at AT TIME ZONE 'UTC';

UPDATE stops
SET departed_at = departed_at + INTERVAL '17 hours'
WHERE departed_at IS NOT NULL
  AND DATE_TRUNC('day', departed_at AT TIME ZONE 'UTC') = departed_at AT TIME ZONE 'UTC';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert: shift T17:00:00Z rows back to T00:00:00Z.
-- Only rows whose time component is exactly 17:00:00 UTC are touched.

UPDATE stops
SET arrived_at = arrived_at - INTERVAL '17 hours'
WHERE EXTRACT(HOUR   FROM arrived_at AT TIME ZONE 'UTC') = 17
  AND EXTRACT(MINUTE FROM arrived_at AT TIME ZONE 'UTC') = 0
  AND EXTRACT(SECOND FROM arrived_at AT TIME ZONE 'UTC') = 0;

UPDATE stops
SET departed_at = departed_at - INTERVAL '17 hours'
WHERE departed_at IS NOT NULL
  AND EXTRACT(HOUR   FROM departed_at AT TIME ZONE 'UTC') = 17
  AND EXTRACT(MINUTE FROM departed_at AT TIME ZONE 'UTC') = 0
  AND EXTRACT(SECOND FROM departed_at AT TIME ZONE 'UTC') = 0;

-- +goose StatementEnd
