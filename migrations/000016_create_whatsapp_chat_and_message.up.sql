-- Criar tabela de chats do whatsapp
CREATE TABLE IF NOT EXISTS whatsapp_chat (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  client_id VARCHAR(36) NOT NULL,
  contact_number VARCHAR(30) NOT NULL,
  contact_name VARCHAR(150) NULL,
  last_message TEXT NULL,
  unread_count INT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_chat (client_id, contact_number),
  FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Criar tabela de mensagens do whatsapp
CREATE TABLE IF NOT EXISTS whatsapp_message (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  chat_id VARCHAR(36) NOT NULL,
  message_id VARCHAR(150) NOT NULL,
  direction ENUM('inbound', 'outbound') NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (chat_id) REFERENCES whatsapp_chat(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
