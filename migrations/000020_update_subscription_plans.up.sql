-- Migration 000020: Atualização dos Planos de Assinatura para Smart e Pro

-- 1. Inserir/Atualizar os novos planos Smart e Pro
INSERT INTO plan (
  id, name, max_professionals, max_customers, max_users, 
  has_loyalty, has_stock, has_reports, has_online_booking, has_whatsapp, 
  is_public, billing_type, price, features_json
) VALUES
(
  'plan-smart', 
  'Smart', 
  3, -1, 3, 
  0, 0, 0, 1, 0, 
  1, 'monthly', 79.90, 
  '{"agenda": true, "caixa": true, "crm_basico": true, "whatsapp": false, "fidelidade": false, "relatorios": false}'
),
(
  'plan-pro', 
  'Pro', 
  -1, -1, -1, 
  1, 1, 1, 1, 1, 
  1, 'monthly', 149.90, 
  '{"agenda": true, "caixa": true, "crm_basico": true, "whatsapp": true, "fidelidade": true, "relatorios": true, "estoque": true}'
),
(
  'p1a00001-smar-47cd-95c5-barber000001', 
  'Smart', 
  3, -1, 3, 
  0, 0, 0, 1, 0, 
  1, 'monthly', 79.90, 
  '{"agenda": true, "caixa": true, "crm_basico": true, "whatsapp": false, "fidelidade": false, "relatorios": false}'
),
(
  'p1a00002-pro0-47cd-95c5-barber000002', 
  'Pro', 
  -1, -1, -1, 
  1, 1, 1, 1, 1, 
  1, 'monthly', 149.90, 
  '{"agenda": true, "caixa": true, "crm_basico": true, "whatsapp": true, "fidelidade": true, "relatorios": true, "estoque": true}'
)
ON DUPLICATE KEY UPDATE 
  name = VALUES(name),
  max_professionals = VALUES(max_professionals),
  max_customers = VALUES(max_customers),
  max_users = VALUES(max_users),
  has_loyalty = VALUES(has_loyalty),
  has_stock = VALUES(has_stock),
  has_reports = VALUES(has_reports),
  has_online_booking = VALUES(has_online_booking),
  has_whatsapp = VALUES(has_whatsapp),
  is_public = VALUES(is_public),
  price = VALUES(price),
  features_json = VALUES(features_json);

-- 2. Atualizar vínculos das barbearias existentes para os novos planos Smart e Pro
UPDATE client
SET plan_id = 'plan-smart'
WHERE plan_id IN ('plan-basico', 'b9a117b3-85b4-47cd-95c5-34c9fb6c25a1');

UPDATE client
SET plan_id = 'plan-pro'
WHERE plan_id IN ('plan-profissional', 'plan-premium', 'c8a227b3-85b4-47cd-95c5-34c9fb6c25a2', 'd7a337b3-85b4-47cd-95c5-34c9fb6c25a3');

-- 3. Remover planos legados descontinuados da base de dados
DELETE FROM plan
WHERE id NOT IN (
  'plan-smart',
  'plan-pro',
  'p1a00001-smar-47cd-95c5-barber000001',
  'p1a00002-pro0-47cd-95c5-barber000002'
);
