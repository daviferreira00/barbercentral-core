-- ============================================
-- ROLLBACK DAS MIGRATIONS DAS FASES 09, 10 E 11
-- ============================================

DROP TABLE IF EXISTS client_block_log;
DROP TABLE IF EXISTS loyalty_transaction;
DROP TABLE IF EXISTS loyalty_card;
DROP TABLE IF EXISTS loyalty_program;

ALTER TABLE plan
  DROP COLUMN max_users,
  DROP COLUMN has_loyalty,
  DROP COLUMN has_stock,
  DROP COLUMN has_reports,
  DROP COLUMN has_online_booking,
  DROP COLUMN is_public;
