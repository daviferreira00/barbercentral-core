-- ============================================
-- CRIAÇÃO DA TABELA client_config
-- ============================================

CREATE TABLE IF NOT EXISTS client_config (
  client_id                 VARCHAR(36) PRIMARY KEY,
  logo_url                  VARCHAR(500) NULL,
  color_primary             VARCHAR(7) NOT NULL DEFAULT '#1a1a1a',
  color_secondary           VARCHAR(7) NOT NULL DEFAULT '#c9a84c',
  font_family               VARCHAR(100) NOT NULL DEFAULT 'Inter',
  address                   VARCHAR(300) NULL,
  neighborhood              VARCHAR(150) NULL,
  city                      VARCHAR(100) NULL,
  state                     VARCHAR(2) NULL,
  phone                     VARCHAR(20) NULL,
  whatsapp                  VARCHAR(20) NULL,
  instagram                 VARCHAR(100) NULL,
  timezone                  VARCHAR(50) NOT NULL DEFAULT 'America/Sao_Paulo',
  cancellation_policy_hours INT NOT NULL DEFAULT 2,
  booking_requires_login    TINYINT(1) NOT NULL DEFAULT 0,
  min_advance_hours         INT NOT NULL DEFAULT 1,
  max_advance_days          INT NOT NULL DEFAULT 30,
  interval_between_minutes  INT NOT NULL DEFAULT 0,
  active                    TINYINT(1) NOT NULL DEFAULT 1,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- ATUALIZAÇÃO DA TABELA appointment
-- ============================================

ALTER TABLE appointment
  ADD COLUMN cancel_token VARCHAR(255) NULL UNIQUE AFTER notes,
  ADD COLUMN customer_name VARCHAR(150) NULL AFTER cancel_token,
  ADD COLUMN customer_phone VARCHAR(20) NULL AFTER customer_name,
  ADD COLUMN customer_email VARCHAR(255) NULL AFTER customer_phone,
  ADD COLUMN reminder_sent TINYINT(1) NOT NULL DEFAULT 0 AFTER customer_email;

-- ============================================
-- SEMENTES (SEEDS)
-- ============================================

INSERT INTO client_config (
  client_id,
  logo_url,
  color_primary,
  color_secondary,
  font_family,
  address,
  neighborhood,
  city,
  state,
  phone,
  whatsapp,
  instagram,
  timezone,
  cancellation_policy_hours,
  booking_requires_login,
  min_advance_hours,
  max_advance_days,
  interval_between_minutes,
  active
) VALUES (
  'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d',
  NULL,
  '#0f172a',
  '#d97706',
  'Inter',
  'Rua das Palmeiras, 150',
  'Centro',
  'São Paulo',
  'SP',
  '11999999999',
  '11999999999',
  'barbeariamodelo',
  'America/Sao_Paulo',
  2,
  0,
  1,
  30,
  15,
  1
) ON DUPLICATE KEY UPDATE color_primary = VALUES(color_primary);
