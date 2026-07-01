-- ============================================
-- 1. CRIAÇÃO DAS TABELAS DA FUNDAÇÃO
-- ============================================

-- Planos
CREATE TABLE IF NOT EXISTS plan (
  id          VARCHAR(36) PRIMARY KEY,
  name        VARCHAR(100) NOT NULL,
  max_professionals INT NOT NULL DEFAULT 5,
  max_customers     INT NOT NULL DEFAULT 500,
  billing_type VARCHAR(20) NOT NULL DEFAULT 'monthly',
  price        DECIMAL(10,2) NOT NULL DEFAULT 0,
  features_json JSON,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Clientes (barbearias)
CREATE TABLE IF NOT EXISTS client (
  id         VARCHAR(36) PRIMARY KEY,
  plan_id    VARCHAR(36) NOT NULL,
  name       VARCHAR(150) NOT NULL,
  slug       VARCHAR(100) NOT NULL UNIQUE,
  status     ENUM('active','blocked') NOT NULL DEFAULT 'active',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (plan_id) REFERENCES plan(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Usuários dos clientes
CREATE TABLE IF NOT EXISTS client_user (
  id            VARCHAR(36) PRIMARY KEY,
  client_id     VARCHAR(36) NOT NULL,
  name          VARCHAR(150) NOT NULL,
  email         VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255),
  role          ENUM('owner','manager','professional','receptionist') NOT NULL DEFAULT 'owner',
  status        ENUM('active','inactive','pending') NOT NULL DEFAULT 'pending',
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES client(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Admin da plataforma
CREATE TABLE IF NOT EXISTS platform_admin (
  id            VARCHAR(36) PRIMARY KEY,
  name          VARCHAR(150) NOT NULL,
  email         VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tokens de autenticação (magic link / reset de senha)
CREATE TABLE IF NOT EXISTS auth_token (
  id         VARCHAR(36) PRIMARY KEY,
  user_id    VARCHAR(36) NOT NULL,
  user_type  ENUM('client_user','platform_admin') NOT NULL,
  type       ENUM('magic_link','password_reset') NOT NULL,
  token      VARCHAR(255) NOT NULL UNIQUE,
  used       TINYINT(1) NOT NULL DEFAULT 0,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 2. SEED DE DADOS INICIAIS
-- ============================================

-- Inserção dos Planos
INSERT INTO plan (id, name, max_professionals, max_customers, billing_type, price, features_json) VALUES
('b9a117b3-85b4-47cd-95c5-34c9fb6c25a1', 'Básico', 2, 200, 'monthly', 0.00, '{"agenda": true, "whatsapp": false}'),
('c8a227b3-85b4-47cd-95c5-34c9fb6c25a2', 'Profissional', 5, 500, 'monthly', 99.00, '{"agenda": true, "whatsapp": true, "financeiro": true}'),
('d7a337b3-85b4-47cd-95c5-34c9fb6c25a3', 'Premium', 9999, 999999, 'monthly', 199.00, '{"agenda": true, "whatsapp": true, "financeiro": true, "fidelidade": true}')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Inserção do Platform Admin inicial (senha: admin123)
INSERT INTO platform_admin (id, name, email, password_hash) VALUES
('a57ce0c7-a4c3-4d6d-88bb-8495679e9514', 'Admin BarberCentral', 'admin@barbercentral.com.br', '$2a$10$16rxE.1lAn2zCCSHX3P6LeVNznY4lqWv6AgW28KyUKtEAjyOymRYK')
ON DUPLICATE KEY UPDATE email = VALUES(email);

-- Inserção da Barbearia de Teste (slug: barbearia-modelo)
INSERT INTO client (id, plan_id, name, slug, status) VALUES
('f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'c8a227b3-85b4-47cd-95c5-34c9fb6c25a2', 'Barbearia Modelo', 'barbearia-modelo', 'active')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Inserção do Usuário Dono (João Barbeiro - senha: senha123)
INSERT INTO client_user (id, client_id, name, email, password_hash, role, status) VALUES
('e2f60b2c-68dc-4e33-91a7-0e625ab73a1e', 'f6e80b2c-68dc-4e33-91a7-0e625ab73a1d', 'João Barbeiro', 'joao@barberiamodelo.com', '$2a$10$NI0.3ybMrRJNdvIpuUVJsegMFhiG/JIByXNVRZYux2L9LN6PfQG9K', 'owner', 'active')
ON DUPLICATE KEY UPDATE email = VALUES(email);
