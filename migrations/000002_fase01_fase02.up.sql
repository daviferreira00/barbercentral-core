-- ============================================
-- 1. CRIAÇÃO DAS TABELAS (FASE 01 & FASE 02)
-- ============================================

-- Profissionais
CREATE TABLE IF NOT EXISTS professional (
  id          VARCHAR(36) PRIMARY KEY,
  client_id   VARCHAR(36) NOT NULL,
  user_id     VARCHAR(36) NULL,
  name        VARCHAR(150) NOT NULL,
  bio         TEXT NULL,
  photo_url   VARCHAR(500) NULL,
  status      ENUM('active','inactive') NOT NULL DEFAULT 'active',
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES client(id),
  FOREIGN KEY (user_id) REFERENCES client_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Categorias de serviços
CREATE TABLE IF NOT EXISTS service_category (
  id        VARCHAR(36) PRIMARY KEY,
  client_id VARCHAR(36) NOT NULL,
  name      VARCHAR(100) NOT NULL,
  FOREIGN KEY (client_id) REFERENCES client(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Serviços
CREATE TABLE IF NOT EXISTS service (
  id                VARCHAR(36) PRIMARY KEY,
  client_id         VARCHAR(36) NOT NULL,
  category_id       VARCHAR(36) NULL,
  name              VARCHAR(150) NOT NULL,
  description       TEXT NULL,
  duration_minutes  INT NOT NULL DEFAULT 30,
  price             DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  photo_url         VARCHAR(500) NULL,
  active            TINYINT(1) NOT NULL DEFAULT 1,
  created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES client(id),
  FOREIGN KEY (category_id) REFERENCES service_category(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Vínculo profissional ↔ serviço
CREATE TABLE IF NOT EXISTS professional_service (
  professional_id   VARCHAR(36) NOT NULL,
  service_id        VARCHAR(36) NOT NULL,
  client_id         VARCHAR(36) NOT NULL,
  custom_price      DECIMAL(10,2) NULL,
  custom_duration   INT NULL,
  PRIMARY KEY (professional_id, service_id),
  FOREIGN KEY (professional_id) REFERENCES professional(id) ON DELETE CASCADE,
  FOREIGN KEY (service_id) REFERENCES service(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Grade de horários semanais
CREATE TABLE IF NOT EXISTS professional_schedule (
  id              VARCHAR(36) PRIMARY KEY,
  professional_id VARCHAR(36) NOT NULL,
  client_id       VARCHAR(36) NOT NULL,
  weekday         TINYINT NOT NULL,
  start_time      TIME NOT NULL,
  end_time        TIME NOT NULL,
  enabled         TINYINT(1) NOT NULL DEFAULT 1,
  UNIQUE KEY uq_prof_weekday (professional_id, weekday),
  FOREIGN KEY (professional_id) REFERENCES professional(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Bloqueios manuais de agenda
CREATE TABLE IF NOT EXISTS blocked_slot (
  id              VARCHAR(36) PRIMARY KEY,
  client_id       VARCHAR(36) NOT NULL,
  professional_id VARCHAR(36) NULL,
  date            DATE NOT NULL,
  start_time      TIME NOT NULL,
  end_time        TIME NOT NULL,
  reason          VARCHAR(255) NULL,
  created_by      VARCHAR(36) NOT NULL,
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES client(id),
  FOREIGN KEY (professional_id) REFERENCES professional(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Agendamentos
CREATE TABLE IF NOT EXISTS appointment (
  id              VARCHAR(36) PRIMARY KEY,
  client_id       VARCHAR(36) NOT NULL,
  professional_id VARCHAR(36) NOT NULL,
  customer_id     VARCHAR(36) NULL,
  date            DATE NOT NULL,
  start_time      TIME NOT NULL,
  end_time        TIME NOT NULL,
  status          ENUM('pending','confirmed','in_progress','completed','cancelled','no_show') NOT NULL DEFAULT 'pending',
  notes           TEXT NULL,
  source          ENUM('panel','online') NOT NULL DEFAULT 'panel',
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES client(id),
  FOREIGN KEY (professional_id) REFERENCES professional(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Serviços do agendamento
CREATE TABLE IF NOT EXISTS appointment_service (
  id               VARCHAR(36) PRIMARY KEY,
  appointment_id   VARCHAR(36) NOT NULL,
  service_id       VARCHAR(36) NOT NULL,
  price            DECIMAL(10,2) NOT NULL,
  duration_minutes INT NOT NULL,
  FOREIGN KEY (appointment_id) REFERENCES appointment(id) ON DELETE CASCADE,
  FOREIGN KEY (service_id) REFERENCES service(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 2. SEED DE DADOS
-- ============================================

-- Inserção de Categorias de Serviços
INSERT INTO service_category (id, client_id, name) VALUES
('cat11111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Cabelo'),
('cat22222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'Barba')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Inserção de Serviços
INSERT INTO service (id, client_id, category_id, name, description, duration_minutes, price, active) VALUES
('srv11111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'cat11111-1111-1111-1111-111111111111', 'Corte Degradê', 'Corte moderno com acabamento em degradê', 30, 45.00, 1),
('srv22222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'cat22222-2222-2222-2222-222222222222', 'Barba Simples', 'Aparar barba com toalha quente e navalha', 20, 30.00, 1),
('srv33333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'cat11111-1111-1111-1111-111111111111', 'Corte + Barba', 'Combo completo corte de cabelo e barba premium', 50, 65.00, 1),
('srv44444-4444-4444-4444-444444444444', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'cat11111-1111-1111-1111-111111111111', 'Hidratação Capilar', 'Tratamento de hidratação e brilho dos fios', 40, 40.00, 1),
('srv55555-5555-5555-5555-555555555555', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'cat22222-2222-2222-2222-222222222222', 'Pigmentação de Barba', 'Camuflagem de fios brancos e preenchimento de falhas', 30, 35.00, 1)
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Inserção de Profissionais
INSERT INTO professional (id, client_id, user_id, name, bio, photo_url, status) VALUES
('p1111111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'e2f60b2c-68dc-4e33-91a7-0e625ab73a1e', 'Marcos Cabeleireiro', 'Especialista em cortes clássicos e tesoura com 10 anos de experiência.', NULL, 'active'),
('p2222222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'Tiago Barbeiro', 'Mestre na navalha e especialista em barboterapia e cortes degradê modernos.', NULL, 'active'),
('p3333333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', NULL, 'Rodolfo Estilista', 'Cabeleireiro com foco em design de estilos, químicas e hidratações.', NULL, 'active')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Inserção de Vínculo de Serviços
INSERT INTO professional_service (professional_id, service_id, client_id) VALUES
('p1111111-1111-1111-1111-111111111111', 'srv11111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d'),
('p1111111-1111-1111-1111-111111111111', 'srv33333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d'),
('p1111111-1111-1111-1111-111111111111', 'srv44444-4444-4444-4444-444444444444', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d'),
('p2222222-2222-2222-2222-222222222222', 'srv11111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d'),
('p2222222-2222-2222-2222-222222222222', 'srv22222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d'),
('p2222222-2222-2222-2222-222222222222', 'srv33333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d'),
('p2222222-2222-2222-2222-222222222222', 'srv55555-5555-5555-5555-555555555555', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d'),
('p3333333-3333-3333-3333-333333333333', 'srv11111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d'),
('p3333333-3333-3333-3333-333333333333', 'srv44444-4444-4444-4444-444444444444', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d')
ON DUPLICATE KEY UPDATE professional_id = VALUES(professional_id);

-- Inserção de Grades de Horários (Seg a Sáb, das 09h às 19h)
INSERT INTO professional_schedule (id, professional_id, client_id, weekday, start_time, end_time, enabled) VALUES
-- Marcos
('s1-1', 'p1111111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 1, '09:00:00', '19:00:00', 1),
('s1-2', 'p1111111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 2, '09:00:00', '19:00:00', 1),
('s1-3', 'p1111111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 3, '09:00:00', '19:00:00', 1),
('s1-4', 'p1111111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 4, '09:00:00', '19:00:00', 1),
('s1-5', 'p1111111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 5, '09:00:00', '19:00:00', 1),
('s1-6', 'p1111111-1111-1111-1111-111111111111', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 6, '09:00:00', '19:00:00', 1),
-- Tiago
('s2-1', 'p2222222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 1, '09:00:00', '19:00:00', 1),
('s2-2', 'p2222222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 2, '09:00:00', '19:00:00', 1),
('s2-3', 'p2222222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 3, '09:00:00', '19:00:00', 1),
('s2-4', 'p2222222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 4, '09:00:00', '19:00:00', 1),
('s2-5', 'p2222222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 5, '09:00:00', '19:00:00', 1),
('s2-6', 'p2222222-2222-2222-2222-222222222222', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 6, '09:00:00', '19:00:00', 1),
-- Rodolfo
('s3-1', 'p3333333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 1, '09:00:00', '19:00:00', 1),
('s3-2', 'p3333333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 2, '09:00:00', '19:00:00', 1),
('s3-3', 'p3333333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 3, '09:00:00', '19:00:00', 1),
('s3-4', 'p3333333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 4, '09:00:00', '19:00:00', 1),
('s3-5', 'p3333333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 5, '09:00:00', '19:00:00', 1),
('s3-6', 'p3333333-3333-3333-3333-333333333333', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 6, '09:00:00', '19:00:00', 1)
ON DUPLICATE KEY UPDATE start_time = VALUES(start_time);

-- Inserção de 1 Bloqueio de Exemplo (Marcos - das 12:00:00 às 13:00:00, hoje)
INSERT INTO blocked_slot (id, client_id, professional_id, date, start_time, end_time, reason, created_by) VALUES
('b1-example', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'p1111111-1111-1111-1111-111111111111', '2026-06-29', '12:00:00', '13:00:00', 'Horário de Almoço', 'e2f60b2c-68dc-4e33-91a7-0e625ab73a1e')
ON DUPLICATE KEY UPDATE reason = VALUES(reason);

-- Inserção de 10 Agendamentos de Exemplo espalhados na semana atual
-- Dia 2026-06-29 (Segunda)
INSERT INTO appointment (id, client_id, professional_id, customer_id, date, start_time, end_time, status, notes, source) VALUES
('ap-1', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'p1111111-1111-1111-1111-111111111111', NULL, '2026-06-29', '10:00:00', '10:30:00', 'completed', 'Deseja corte tesoura clássico', 'panel'),
('ap-2', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'p2222222-2222-2222-2222-222222222222', NULL, '2026-06-29', '11:00:00', '11:20:00', 'completed', 'Barba toalha quente', 'panel'),
('ap-3', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'p1111111-1111-1111-1111-111111111111', NULL, '2026-06-29', '14:00:00', '14:50:00', 'confirmed', 'Corte completo e barba', 'panel'),
('ap-4', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'p3333333-3333-3333-3333-333333333333', NULL, '2026-06-29', '15:00:00', '15:40:00', 'in_progress', 'Tratamento de hidratação', 'panel'),
('ap-5', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'p2222222-2222-2222-2222-222222222222', NULL, '2026-06-29', '16:00:00', '16:30:00', 'pending', 'Corte degradê de máquina', 'panel'),
-- Dia 2026-06-30 (Terça)
('ap-6', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'p1111111-1111-1111-1111-111111111111', NULL, '2026-06-30', '09:30:00', '10:00:00', 'confirmed', 'Corte degradê', 'panel'),
('ap-7', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'p2222222-2222-2222-2222-222222222222', NULL, '2026-06-30', '10:30:00', '11:20:00', 'confirmed', 'Corte e barba combo', 'panel'),
('ap-8', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'p3333333-3333-3333-3333-333333333333', NULL, '2026-06-30', '14:00:00', '14:30:00', 'pending', 'Corte degradê', 'panel'),
-- Dia 2026-07-01 (Quarta)
('ap-9', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'p2222222-2222-2222-2222-222222222222', NULL, '2026-07-01', '15:00:00', '15:30:00', 'confirmed', 'Camuflagem e preenchimento', 'panel'),
('ap-10', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'p1111111-1111-1111-1111-111111111111', NULL, '2026-07-01', '16:00:00', '16:40:00', 'confirmed', 'Hidratação rápida', 'panel')
ON DUPLICATE KEY UPDATE date = VALUES(date);

-- Inserção dos Serviços vinculados aos Agendamentos
INSERT INTO appointment_service (id, appointment_id, service_id, price, duration_minutes) VALUES
('aps-1', 'ap-1', 'srv11111-1111-1111-1111-111111111111', 45.00, 30),
('aps-2', 'ap-2', 'srv22222-2222-2222-2222-222222222222', 30.00, 20),
('aps-3', 'ap-3', 'srv33333-3333-3333-3333-333333333333', 65.00, 50),
('aps-4', 'ap-4', 'srv44444-4444-4444-4444-444444444444', 40.00, 40),
('aps-5', 'ap-5', 'srv11111-1111-1111-1111-111111111111', 45.00, 30),
('aps-6', 'ap-6', 'srv11111-1111-1111-1111-111111111111', 45.00, 30),
('aps-7', 'ap-7', 'srv33333-3333-3333-3333-333333333333', 65.00, 50),
('aps-8', 'ap-8', 'srv11111-1111-1111-1111-111111111111', 45.00, 30),
('aps-9', 'ap-9', 'srv55555-5555-5555-5555-555555555555', 35.00, 30),
('aps-10', 'ap-10', 'srv44444-4444-4444-4444-444444444444', 40.00, 40)
ON DUPLICATE KEY UPDATE price = VALUES(price);
