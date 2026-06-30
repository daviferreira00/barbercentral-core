-- ============================================
-- 1. CRIAÇÃO DA TABELA customer (CRM)
-- ============================================

CREATE TABLE IF NOT EXISTS customer (
  id          VARCHAR(36) PRIMARY KEY,
  client_id   VARCHAR(36) NOT NULL,
  name        VARCHAR(150) NOT NULL,
  phone       VARCHAR(20) NOT NULL,
  email       VARCHAR(255) NULL,
  cpf         VARCHAR(14) NULL,
  birth_date  DATE NULL,
  notes       TEXT NULL,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uq_customer_phone (client_id, phone),
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 2. VINCULAR customer_id NA TABELA appointment
-- ============================================

ALTER TABLE appointment
  ADD CONSTRAINT fk_appointment_customer FOREIGN KEY (customer_id) REFERENCES customer(id) ON DELETE SET NULL;

-- ============================================
-- 3. CRIAÇÃO DA TABELA appointment_payment (BASE)
-- ============================================

CREATE TABLE IF NOT EXISTS appointment_payment (
  id             VARCHAR(36) PRIMARY KEY,
  appointment_id VARCHAR(36) NOT NULL UNIQUE,
  client_id      VARCHAR(36) NOT NULL,
  amount         DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  method         ENUM('cash','pix','card_debit','card_credit','other') NOT NULL DEFAULT 'cash',
  status         ENUM('pending','paid','refunded') NOT NULL DEFAULT 'pending',
  paid_at        DATETIME NULL,
  notes          VARCHAR(300) NULL,
  created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (appointment_id) REFERENCES appointment(id) ON DELETE CASCADE,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 4. CRIAÇÃO DA TABELA appointment_status_log
-- ============================================

CREATE TABLE IF NOT EXISTS appointment_status_log (
  id             VARCHAR(36) PRIMARY KEY,
  appointment_id VARCHAR(36) NOT NULL,
  from_status    VARCHAR(30) NULL,
  to_status      VARCHAR(30) NOT NULL,
  changed_by     VARCHAR(36) NOT NULL, -- UUID de client_user
  notes          TEXT NULL,
  created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (appointment_id) REFERENCES appointment(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 5. POPULAR SEMENTES (SEEDS) COM 15 CLIENTES
-- ============================================

INSERT INTO customer (id, client_id, name, phone, email, cpf, birth_date, notes) VALUES
('c1111111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Pedro Alvares', '11911111111', 'pedro@email.com', '123.456.789-01', '1990-05-15', 'Cliente VIP, prefere café expresso.'),
('c2222222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Lucas Silva', '11922222222', 'lucas@email.com', '234.567.890-12', '1985-11-20', 'Cabelo afro, costuma fazer degradê na zero.'),
('c3333333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Mariana Costa', '11933333333', 'mariana@email.com', NULL, '1998-02-28', 'Faz alinhamento capilar bimestral.'),
('c4444444-4444-4444-4444-444444444444', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Carlos Eduardo', '11944444444', 'carlos@email.com', '345.678.901-23', '1979-07-04', 'Sempre pontual.'),
('c5555555-5555-5555-5555-555555555555', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Gabriel Souza', '11955555555', 'gabriel@email.com', NULL, '1993-10-12', 'Vem a cada 15 dias.'),
('c6666666-6666-6666-6666-666666666666', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Bruno Henrique', '11966666666', 'bruno@email.com', '456.789.012-34', '1991-03-08', NULL),
('c7777777-7777-7777-7777-777777777777', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Felipe Santos', '11977777777', 'felipe@email.com', NULL, '1995-12-25', 'Prefere cortes clássicos na tesoura.'),
('c8888888-8888-8888-8888-888888888888', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Thiago Lima', '11988888888', 'thiago@email.com', '567.890.123-45', '1988-08-30', NULL),
('c9999999-9999-9999-9999-999999999999', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Rodrigo Alves', '11999999990', 'rodrigo@email.com', NULL, '1994-06-18', NULL),
('c1010101-1010-1010-1010-101010101010', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Matheus Oliveira', '11910101010', 'matheus@email.com', '678.901.234-56', '1987-04-02', 'Corte e barba terapia.'),
('c1101101-1101-1101-1101-110110110110', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Daniel Pires', '11911011101', 'daniel@email.com', NULL, '1992-09-09', NULL),
('c1201201-1201-1201-1201-120120120120', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Rafael Dutra', '11912011201', 'rafael@email.com', '789.012.345-67', '2000-01-01', NULL),
('c1301301-1301-1301-1301-130130130130', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Alexandre Borges', '11913011301', 'alexandre@email.com', NULL, '1983-05-30', NULL),
('c1401401-1401-1401-1401-140140140140', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Leonardo Cruz', '11914011401', 'leonardo@email.com', '890.123.456-78', '1996-08-12', NULL),
('c1501501-1501-1501-1501-150150150150', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Gustavo Paiva', '11915011501', 'gustavo@email.com', NULL, '1997-12-03', NULL)
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Vincular clientes aos agendamentos antigos para simular histórico completo
UPDATE appointment SET customer_id = 'c1111111-1111-1111-1111-111111111111' WHERE id IN ('ap-1', 'ap-6', 'ap-10');
UPDATE appointment SET customer_id = 'c2222222-2222-2222-2222-222222222222' WHERE id IN ('ap-2', 'ap-7');
UPDATE appointment SET customer_id = 'c3333333-3333-3333-3333-333333333333' WHERE id IN ('ap-4', 'ap-8');

-- Inserir logs de status retroativos
INSERT INTO appointment_status_log (id, appointment_id, from_status, to_status, changed_by, notes) VALUES
('l-1', 'ap-1', 'pending', 'confirmed', 'e2f60b2c-68dc-4e33-91a7-0e625ab73a1e', 'Agendamento aceito pelo profissional'),
('l-2', 'ap-1', 'confirmed', 'completed', 'e2f60b2c-68dc-4e33-91a7-0e625ab73a1e', 'Serviço concluído na cadeira'),
('l-3', 'ap-2', 'pending', 'completed', 'e2f60b2c-68dc-4e33-91a7-0e625ab73a1e', 'Atendimento rápido direto para conclusão'),
('l-4', 'ap-3', 'pending', 'confirmed', 'e2f60b2c-68dc-4e33-91a7-0e625ab73a1e', 'Confirmado pelo painel administrativo')
ON DUPLICATE KEY UPDATE to_status = VALUES(to_status);
