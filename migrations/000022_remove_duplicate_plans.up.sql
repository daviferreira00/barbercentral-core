-- Migration 000022: Remover planos duplicados legados Smart e Pro

-- 1. Redirecionar quaisquer barbearias que ainda estejam associadas aos IDs duplicados/legados
UPDATE client 
SET plan_id = 'plan-smart' 
WHERE plan_id IN ('p1a00001-smar-47cd-95c5-barber000001', 'p1a00001-smart-47cd-95c5-barber000001');

UPDATE client 
SET plan_id = 'plan-pro' 
WHERE plan_id IN ('p1a00002-pro0-47cd-95c5-barber000002', 'p1a00002-pro000-47cd-95c5-barber000002');

-- 2. Remover os registros duplicados de plan, mantendo apenas os canônicos 'plan-smart' e 'plan-pro'
DELETE FROM plan 
WHERE id IN (
  'p1a00001-smar-47cd-95c5-barber000001',
  'p1a00001-smart-47cd-95c5-barber000001',
  'p1a00002-pro0-47cd-95c5-barber000002',
  'p1a00002-pro000-47cd-95c5-barber000002'
);
