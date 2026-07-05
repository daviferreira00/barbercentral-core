-- ============================================
-- Multi-tenant por usuário: separa a identidade da conta (user_account)
-- do vínculo com uma barbearia específica (client_user_link).
-- Antes: client_user tinha client_id/role fixos (1 usuário = 1 barbearia).
-- Depois: um mesmo usuário pode ter N vínculos ativos (um por barbearia),
-- cada vínculo com seu próprio role/status.
-- ============================================

CREATE TABLE IF NOT EXISTS client_user_link (
  id         VARCHAR(36) PRIMARY KEY,
  user_id    VARCHAR(36) NOT NULL,
  client_id  VARCHAR(36) NOT NULL,
  role       ENUM('owner','manager','professional','receptionist') NOT NULL DEFAULT 'owner',
  status     ENUM('active','inactive','pending') NOT NULL DEFAULT 'pending',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_user_client (user_id, client_id),
  FOREIGN KEY (user_id) REFERENCES client_user(id) ON DELETE CASCADE,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Copia o vínculo único que já existe hoje (1 usuário = 1 barbearia) para a nova tabela
INSERT INTO client_user_link (id, user_id, client_id, role, status, created_at)
SELECT UUID(), id, client_id, role, status, created_at FROM client_user;

-- Renomeia client_user -> user_account, preservando os MESMOS ids
-- (a FK de professional.user_id -> client_user(id) sobrevive automaticamente ao rename)
ALTER TABLE client_user RENAME TO user_account;

-- Remove a FK de client_id (nome auto-gerado na criação original; descoberto dinamicamente
-- para não depender do nome exato atribuído pelo MySQL/MariaDB)
SET @fk_name = (
  SELECT CONSTRAINT_NAME FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_account'
    AND COLUMN_NAME = 'client_id'
    AND REFERENCED_TABLE_NAME = 'client'
  LIMIT 1
);
SET @drop_fk_sql = CONCAT('ALTER TABLE user_account DROP FOREIGN KEY ', @fk_name);
PREPARE stmt FROM @drop_fk_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- client_id/role agora vivem em client_user_link (por vínculo); user_account.status
-- passa a significar "consegue logar em algum lugar da plataforma" (status GLOBAL da conta)
ALTER TABLE user_account DROP COLUMN client_id;
ALTER TABLE user_account DROP COLUMN role;
