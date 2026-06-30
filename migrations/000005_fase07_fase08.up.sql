-- ============================================
-- 1. CRIAÇÃO DAS TABELAS FINANCEIRAS
-- ============================================

CREATE TABLE IF NOT EXISTS cash_register (
  id              VARCHAR(36) PRIMARY KEY,
  client_id       VARCHAR(36) NOT NULL,
  opened_by       VARCHAR(36) NOT NULL,
  opened_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  closed_by       VARCHAR(36) NULL,
  closed_at       DATETIME NULL,
  opening_balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  closing_balance DECIMAL(10,2) NULL,
  notes           TEXT NULL,
  status          ENUM('open','closed') NOT NULL DEFAULT 'open',
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cash_transaction (
  id                     VARCHAR(36) PRIMARY KEY,
  register_id            VARCHAR(36) NOT NULL,
  client_id              VARCHAR(36) NOT NULL,
  appointment_payment_id VARCHAR(36) NULL,
  type                   ENUM('income','expense') NOT NULL,
  amount                 DECIMAL(10,2) NOT NULL,
  method                 ENUM('cash','pix','card_debit','card_credit','other') NOT NULL DEFAULT 'cash',
  description            VARCHAR(300) NOT NULL,
  category               VARCHAR(100) NULL,
  created_by             VARCHAR(36) NOT NULL,
  created_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (register_id) REFERENCES cash_register(id) ON DELETE CASCADE,
  FOREIGN KEY (appointment_payment_id) REFERENCES appointment_payment(id) ON DELETE SET NULL,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 2. CRIAÇÃO DAS TABELAS DE ESTOQUE
-- ============================================

CREATE TABLE IF NOT EXISTS product (
  id               VARCHAR(36) PRIMARY KEY,
  client_id        VARCHAR(36) NOT NULL,
  name             VARCHAR(150) NOT NULL,
  description      TEXT NULL,
  sku              VARCHAR(50) NULL,
  price            DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  cost_price       DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  quantity_in_stock DECIMAL(10,3) NOT NULL DEFAULT 0.000,
  low_stock_alert  DECIMAL(10,3) NOT NULL DEFAULT 5.000,
  unit             VARCHAR(20) NOT NULL DEFAULT 'un',
  active           TINYINT(1) NOT NULL DEFAULT 1,
  created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_movement (
  id             VARCHAR(36) PRIMARY KEY,
  product_id     VARCHAR(36) NOT NULL,
  client_id      VARCHAR(36) NOT NULL,
  type           ENUM('in','out','adjustment') NOT NULL,
  quantity       DECIMAL(10,3) NOT NULL,
  reason         VARCHAR(300) NULL,
  appointment_id VARCHAR(36) NULL,
  created_by     VARCHAR(36) NOT NULL,
  created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (product_id) REFERENCES product(id) ON DELETE CASCADE,
  FOREIGN KEY (appointment_id) REFERENCES appointment(id) ON DELETE SET NULL,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS service_product (
  service_id  VARCHAR(36) NOT NULL,
  product_id  VARCHAR(36) NOT NULL,
  client_id   VARCHAR(36) NOT NULL,
  quantity    DECIMAL(10,3) NOT NULL DEFAULT 1.000,
  PRIMARY KEY (service_id, product_id),
  FOREIGN KEY (service_id) REFERENCES service(id) ON DELETE CASCADE,
  FOREIGN KEY (product_id) REFERENCES product(id) ON DELETE CASCADE,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 3. SEMENTES DE CAIXAS (SEEDS)
-- ============================================

INSERT INTO cash_register (id, client_id, opened_by, opened_at, closed_by, closed_at, opening_balance, closing_balance, notes, status) VALUES
('cr-yesterday', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'u2222222-2222-2222-2222-222222222222', '2026-06-28 08:00:00', 'u2222222-2222-2222-2222-222222222222', '2026-06-28 20:00:00', 100.00, 450.00, 'Caixa fechado ontem sem divergências', 'closed'),
('cr-today', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'u2222222-2222-2222-2222-222222222222', '2026-06-29 08:00:00', NULL, NULL, 150.00, NULL, NULL, 'open')
ON DUPLICATE KEY UPDATE status = VALUES(status);

-- ============================================
-- 4. SEMENTES DE TRANSAÇÕES FINANCEIRAS
-- ============================================

INSERT INTO cash_transaction (id, register_id, client_id, appointment_payment_id, type, amount, method, description, category, created_by, created_at) VALUES
('tx-1', 'cr-yesterday', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'income', 80.00, 'cash', 'Lançamento manual: Venda Pomada Modeladora', 'Vendas', 'u2222222-2222-2222-2222-222222222222', '2026-06-28 09:30:00'),
('tx-2', 'cr-yesterday', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'expense', 30.00, 'cash', 'Lançamento manual: Compra pó de café', 'Suprimentos', 'u2222222-2222-2222-2222-222222222222', '2026-06-28 10:15:00'),
('tx-3', 'cr-yesterday', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'income', 150.00, 'pix', 'Venda combo Shampoo e Condicionador', 'Vendas', 'u2222222-2222-2222-2222-222222222222', '2026-06-28 14:00:00'),
('tx-4', 'cr-yesterday', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'expense', 15.00, 'cash', 'Lançamento manual: Compra água mineral', 'Copa', 'u2222222-2222-2222-2222-222222222222', '2026-06-28 15:30:00'),
('tx-5', 'cr-yesterday', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'income', 165.00, 'card_credit', 'Recebimento Serviços do Dia', 'Serviços', 'u2222222-2222-2222-2222-222222222222', '2026-06-28 18:00:00'),
('tx-6', 'cr-today', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'income', 45.00, 'pix', 'Venda Óleo para Barba Club', 'Vendas', 'u2222222-2222-2222-2222-222222222222', '2026-06-29 09:10:00'),
('tx-7', 'cr-today', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'expense', 50.00, 'cash', 'Lançamento manual: Material de Limpeza', 'Manutenção', 'u2222222-2222-2222-2222-222222222222', '2026-06-29 11:45:00'),
('tx-8', 'cr-today', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'income', 120.00, 'card_debit', 'Recebimento de Corte e Barboterapia', 'Serviços', 'u2222222-2222-2222-2222-222222222222', '2026-06-29 14:30:00')
ON DUPLICATE KEY UPDATE register_id = VALUES(register_id);

-- ============================================
-- 5. SEMENTES DE PRODUTOS
-- ============================================

INSERT INTO product (id, client_id, name, description, sku, price, cost_price, quantity_in_stock, low_stock_alert, unit, active) VALUES
('p1', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Pomada Modeladora Matte Club', 'Pomada para cabelo efeito fosco e alta fixação', 'PM-MATTE-01', 59.90, 25.00, 15.000, 3.000, 'un', 1),
('p2', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Shampoo Anticaspa Ice 250ml', 'Shampoo refrescante de uso diário', 'SH-ICE-250', 39.90, 18.00, 2.000, 5.000, 'un', 1), -- ESTOQUE BAIXO (2 < 5)
('p3', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Condicionador Hidratante 250ml', 'Condicionador para todos os tipos de cabelo', 'COND-HID-250', 42.90, 19.50, 12.000, 4.000, 'un', 1),
('p4', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Óleo para Barba Premium 30ml', 'Óleo hidratante e perfumado para barba', 'OL-BARBA-30', 49.90, 22.00, 0.000, 2.000, 'un', 1), -- SEM ESTOQUE (0 < 2)
('p5', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Goma de Barbear Refresh 100g', 'Facilita o deslizar da lâmina de barbear', 'GOM-BAR-100', 34.90, 15.00, 8.000, 3.000, 'un', 1),
('p6', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Tônico Capilar Crescimento', 'Tônico para fortalecimento dos fios', 'TON-CRESC-50', 89.90, 38.00, 6.000, 2.000, 'un', 1),
('p7', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Cerveja Heineken Long Neck', 'Cerveja Premium Lager Pilsen', 'CERV-HEIN-LN', 10.00, 4.50, 24.000, 10.000, 'un', 1),
('p8', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Cerveja Stella Artois LN', 'Cerveja Premium Pilsen Belga', 'CERV-STELLA-LN', 10.00, 4.20, 18.000, 10.000, 'un', 1),
('p9', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Refrigerante Coca-Cola Lata', 'Lata 350ml tradicional', 'REF-COCA-LATA', 6.00, 2.20, 30.000, 8.000, 'un', 1),
('p10', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Café Gourmet Espresso', 'Grãos selecionados moídos na hora', 'CAF-ESP-01', 5.00, 1.20, 4.500, 1.000, 'kg', 1)
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- ============================================
-- 6. SEMENTES DE MOVIMENTAÇÕES DE ESTOQUE
-- ============================================

INSERT INTO stock_movement (id, product_id, client_id, type, quantity, reason, appointment_id, created_by, created_at) VALUES
('m-1', 'p1', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'in', 20.000, 'Compra inicial de lote do fornecedor', NULL, 'u2222222-2222-2222-2222-222222222222', '2026-06-28 08:30:00'),
('m-2', 'p1', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'out', 5.000, 'Venda direta balcão para cliente', NULL, 'u2222222-2222-2222-2222-222222222222', '2026-06-28 14:00:00'),
('m-3', 'p2', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'in', 10.000, 'Entrada de estoque de reposição', NULL, 'u2222222-2222-2222-2222-222222222222', '2026-06-28 08:30:00'),
('m-4', 'p2', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'out', 8.000, 'Ajuste por vazamento/quebra de frascos', NULL, 'u2222222-2222-2222-2222-222222222222', '2026-06-28 17:30:00'),
('m-5', 'p4', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'in', 5.000, 'Compra inicial para testes', NULL, 'u2222222-2222-2222-2222-222222222222', '2026-06-28 08:30:00'),
('m-6', 'p4', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'adjustment', 0.000, 'Zerar estoque físico para balanço geral', NULL, 'u2222222-2222-2222-2222-222222222222', '2026-06-29 10:00:00')
ON DUPLICATE KEY UPDATE product_id = VALUES(product_id);

-- ============================================
-- 7. VÍNCULO DE CONSUMO AUTOMÁTICO
-- ============================================

INSERT INTO service_product (service_id, product_id, client_id, quantity) VALUES
('srv11111-1111-1111-1111-111111111111', 'p2', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 0.050), -- Corte de Cabelo consome 0.05 unidades de Shampoo (ex: fração)
('srv22222-2222-2222-2222-222222222222', 'p5', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 0.100)  -- Barba consome 0.1 unidades de Goma de Barbear
ON DUPLICATE KEY UPDATE quantity = VALUES(quantity);
