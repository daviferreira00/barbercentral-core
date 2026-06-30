-- ============================================
-- MIGRATIONS DAS FASES 09, 10 E 11 (UP)
-- ============================================

-- 1. ALTERAÇÕES NA TABELA PLAN
ALTER TABLE plan
  ADD COLUMN max_users INT NOT NULL DEFAULT 3 AFTER max_customers,
  ADD COLUMN has_loyalty TINYINT(1) NOT NULL DEFAULT 0 AFTER max_users,
  ADD COLUMN has_stock TINYINT(1) NOT NULL DEFAULT 0 AFTER has_loyalty,
  ADD COLUMN has_reports TINYINT(1) NOT NULL DEFAULT 0 AFTER has_stock,
  ADD COLUMN has_online_booking TINYINT(1) NOT NULL DEFAULT 1 AFTER has_reports,
  ADD COLUMN is_public TINYINT(1) NOT NULL DEFAULT 1 AFTER has_online_booking;

-- 2. TABELAS DE FIDELIDADE (FASE 09)
CREATE TABLE IF NOT EXISTS loyalty_program (
  id                   VARCHAR(36) PRIMARY KEY,
  client_id            VARCHAR(36) NOT NULL UNIQUE,
  name                 VARCHAR(150) NOT NULL,
  type                 ENUM('stamps','points') NOT NULL,
  stamps_to_reward     INT,          -- tipo stamps: quantos carimbos = 1 recompensa
  points_per_real      DECIMAL(5,2), -- tipo points: pontos por R$ 1,00 gasto
  reward_description   TEXT NOT NULL,
  active               TINYINT(1) NOT NULL DEFAULT 1,
  created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS loyalty_card (
  id             VARCHAR(36) PRIMARY KEY,
  customer_id    VARCHAR(36) NOT NULL,
  client_id      VARCHAR(36) NOT NULL,
  program_id     VARCHAR(36) NOT NULL,
  stamps_count   INT NOT NULL DEFAULT 0,
  points_balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  status         ENUM('active','inactive') NOT NULL DEFAULT 'active',
  created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uq_card_customer_client (customer_id, client_id),
  FOREIGN KEY (customer_id) REFERENCES customer(id) ON DELETE CASCADE,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE,
  FOREIGN KEY (program_id) REFERENCES loyalty_program(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS loyalty_transaction (
  id             VARCHAR(36) PRIMARY KEY,
  card_id        VARCHAR(36) NOT NULL,
  client_id      VARCHAR(36) NOT NULL,
  appointment_id VARCHAR(36) NULL,
  type           ENUM('earn','redeem') NOT NULL,
  stamps_value   INT NULL,
  points_value   DECIMAL(10,2) NULL,
  description    VARCHAR(300) NOT NULL,
  created_by     VARCHAR(36) NOT NULL,
  created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (card_id) REFERENCES loyalty_card(id) ON DELETE CASCADE,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE,
  FOREIGN KEY (appointment_id) REFERENCES appointment(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. AUDITORIA DE BLOQUEIOS (FASE 11)
CREATE TABLE IF NOT EXISTS client_block_log (
  id           VARCHAR(36) PRIMARY KEY,
  client_id    VARCHAR(36) NOT NULL,
  action       ENUM('block','unblock') NOT NULL,
  reason       VARCHAR(500) NULL,
  performed_by VARCHAR(36) NOT NULL,
  created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- SEMENTES E ATUALIZAÇÕES DE DADOS (SEEDS)
-- ============================================

-- Atualizar limites dos planos existentes
UPDATE plan SET max_users = 3, has_loyalty = 0, has_stock = 0, has_reports = 0, has_online_booking = 1, is_public = 1 WHERE id = 'b9a117b3-85b4-47cd-95c5-34c9fb6c25a1';
UPDATE plan SET max_users = 5, has_loyalty = 1, has_stock = 1, has_reports = 1, has_online_booking = 1, is_public = 1 WHERE id = 'c8a227b3-85b4-47cd-95c5-34c9fb6c25a2';
UPDATE plan SET max_users = 9999, has_loyalty = 1, has_stock = 1, has_reports = 1, has_online_booking = 1, is_public = 1 WHERE id = 'd7a337b3-85b4-47cd-95c5-34c9fb6c25a3';

-- Inserir Programa de Fidelidade Ativo para a barbearia modelo
INSERT INTO loyalty_program (id, client_id, name, type, stamps_to_reward, points_per_real, reward_description, active) VALUES
('lp-modelo', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Fidelidade Barber Club', 'stamps', 10, NULL, 'Corte de cabelo ou barba grátis', 1)
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Cartões para clientes fictícios do seed (Fase 04)
INSERT INTO loyalty_card (id, customer_id, client_id, program_id, stamps_count, points_balance, status) VALUES
('lc-1', 'c1111111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'lp-modelo', 4, 0.00, 'active'),
('lc-2', 'c2222222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'lp-modelo', 9, 0.00, 'active'), -- QUASE PREMIADO (9/10)
('lc-3', 'c3333333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'lp-modelo', 0, 0.00, 'active'),
('lc-4', 'c4444444-4444-4444-4444-444444444444', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'lp-modelo', 7, 0.00, 'active')
ON DUPLICATE KEY UPDATE stamps_count = VALUES(stamps_count);

-- Histórico de transações
INSERT INTO loyalty_transaction (id, card_id, client_id, appointment_id, type, stamps_value, points_value, description, created_by, created_at) VALUES
('ltx-1', 'lc-1', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'earn', 4, 0.00, 'Saldo inicial migrado', 'system', '2026-06-28 09:00:00'),
('ltx-2', 'lc-2', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'earn', 9, 0.00, 'Acúmulo de atendimentos passados', 'system', '2026-06-28 10:00:00'),
('ltx-3', 'lc-4', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'earn', 7, 0.00, 'Acúmulo de atendimentos passados', 'system', '2026-06-28 11:00:00')
ON DUPLICATE KEY UPDATE stamps_value = VALUES(stamps_value);
