ALTER TABLE client_config
  ALTER COLUMN max_advance_days SET DEFAULT 30;

UPDATE client_config
SET max_advance_days = 30
WHERE max_advance_days = 60;
