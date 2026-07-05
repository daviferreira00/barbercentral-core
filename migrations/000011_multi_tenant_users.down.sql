-- ATENÇÃO: rollback com perda de dados se, no momento em que este down.sql rodar,
-- algum usuário já tiver 2+ vínculos ativos em client_user_link — só o PRIMEIRO
-- vínculo encontrado por usuário é restaurado em user_account/client_user
-- (não há como um usuário voltar a ser 1:1 sem escolher qual vínculo descartar).

ALTER TABLE user_account ADD COLUMN client_id VARCHAR(36) NULL AFTER id;
ALTER TABLE user_account ADD COLUMN role ENUM('owner','manager','professional','receptionist') NOT NULL DEFAULT 'owner' AFTER password_hash;

UPDATE user_account ua
JOIN (
  SELECT user_id, MIN(id) AS link_id
  FROM client_user_link
  WHERE status = 'active'
  GROUP BY user_id
) first_link ON first_link.user_id = ua.id
JOIN client_user_link cul ON cul.id = first_link.link_id
SET ua.client_id = cul.client_id, ua.role = cul.role;

DELETE FROM user_account WHERE client_id IS NULL;

ALTER TABLE user_account MODIFY COLUMN client_id VARCHAR(36) NOT NULL;
ALTER TABLE user_account ADD CONSTRAINT user_account_client_fk FOREIGN KEY (client_id) REFERENCES client(id);

DROP TABLE IF EXISTS client_user_link;

ALTER TABLE user_account RENAME TO client_user;
