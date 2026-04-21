-- Drop zoom_accounts table
DROP INDEX IF EXISTS idx_zoom_accounts_user_active;
DROP INDEX IF EXISTS idx_zoom_accounts_deleted_at;
DROP INDEX IF EXISTS idx_zoom_accounts_is_active;
DROP INDEX IF EXISTS idx_zoom_accounts_user_id;

DROP TABLE IF EXISTS zoom_accounts;