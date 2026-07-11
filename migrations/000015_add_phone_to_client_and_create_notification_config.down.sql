-- 1. Remover tabela de configuração de notificações
DROP TABLE IF EXISTS notification_config;

-- 2. Remover campo phone da tabela client
ALTER TABLE client DROP COLUMN phone;
