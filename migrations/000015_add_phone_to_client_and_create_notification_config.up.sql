-- 1. Adicionar campo phone na tabela client
ALTER TABLE client ADD COLUMN phone VARCHAR(20) NULL;

-- 2. Criar tabela de configuração de notificações
CREATE TABLE IF NOT EXISTS notification_config (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  client_id VARCHAR(36) NOT NULL,
  name VARCHAR(100) NOT NULL,
  trigger_type ENUM('booking_confirmation', 'booking_reminder', 'customer_retention') NOT NULL,
  trigger_value INT NOT NULL,
  trigger_unit ENUM('hours', 'days') NOT NULL,
  message_template TEXT NOT NULL,
  channel_id VARCHAR(36) NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE,
  FOREIGN KEY (channel_id) REFERENCES whatsapp_instance(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
