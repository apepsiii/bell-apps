-- Migration: Add Gowa WhatsApp Gateway settings
-- Date: 2026-09-16

-- Add username and device_id settings for Gowa integration
INSERT OR IGNORE INTO attendance_settings (setting_key, setting_value) 
VALUES ('onesender_username', '');

INSERT OR IGNORE INTO attendance_settings (setting_key, setting_value) 
VALUES ('onesender_device_id', '');
