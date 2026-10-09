-- ============================================
-- Migration 000021: Trial de 7 dias e Onboarding/Tutoriais
-- ============================================

ALTER TABLE client 
  ADD COLUMN trial_ends_at DATETIME NULL AFTER status,
  ADD COLUMN subscription_status ENUM('trial', 'active', 'past_due', 'canceled', 'blocked') NOT NULL DEFAULT 'trial' AFTER trial_ends_at;

-- Barbearias já existentes permanecem ativas sem expiracao de trial
UPDATE client SET subscription_status = 'active' WHERE status = 'active';

ALTER TABLE client_user_link 
  ADD COLUMN completed_onboarding TINYINT(1) NOT NULL DEFAULT 0 AFTER status,
  ADD COLUMN seen_tutorials_json JSON NULL AFTER completed_onboarding;
