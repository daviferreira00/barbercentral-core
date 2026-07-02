-- Adicionar coluna custom_domain na tabela client
ALTER TABLE client ADD COLUMN custom_domain VARCHAR(255) NULL UNIQUE;
