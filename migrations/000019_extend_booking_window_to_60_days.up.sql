ALTER TABLE client_config
  ALTER COLUMN max_advance_days SET DEFAULT 60;

UPDATE client_config
SET max_advance_days = 60
WHERE max_advance_days < 60;
