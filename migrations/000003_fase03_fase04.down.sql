-- ============================================
-- REVERSÃO DA MIGRATION FASE 03 & FASE 04
-- ============================================

DROP TABLE IF EXISTS client_config;

ALTER TABLE appointment
  DROP COLUMN IF EXISTS reminder_sent,
  DROP COLUMN IF EXISTS customer_email,
  DROP COLUMN IF EXISTS customer_phone,
  DROP COLUMN IF EXISTS customer_name,
  DROP COLUMN IF EXISTS cancel_token;
