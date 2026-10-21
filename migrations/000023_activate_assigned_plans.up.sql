-- ==========================================================
-- Migration 000023: Ativar barbearias que já tiveram plano definido
-- ==========================================================

-- Ativa barbearias com status 'active' cujo trial já expirou
UPDATE client 
SET subscription_status = 'active' 
WHERE status = 'active' 
  AND subscription_status = 'trial' 
  AND trial_ends_at IS NOT NULL 
  AND trial_ends_at < NOW();
