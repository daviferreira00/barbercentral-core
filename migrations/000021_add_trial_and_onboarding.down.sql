-- Migration 000021 Down

ALTER TABLE client_user_link 
  DROP COLUMN seen_tutorials_json,
  DROP COLUMN completed_onboarding;

ALTER TABLE client 
  DROP COLUMN subscription_status,
  DROP COLUMN trial_ends_at;
