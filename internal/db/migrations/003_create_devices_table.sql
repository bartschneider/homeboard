-- Migration: Create devices table for KUAL extension support
-- This table stores registered Kindle devices and their dashboard assignments

CREATE TABLE IF NOT EXISTS devices (
    device_id TEXT PRIMARY KEY NOT NULL,
    device_name TEXT NOT NULL,
    device_type TEXT NOT NULL DEFAULT 'kindle',
    capabilities TEXT NOT NULL DEFAULT '[]', -- JSON array of device capabilities
    dashboard_url TEXT DEFAULT '',
    last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_devices_type ON devices(device_type);
CREATE INDEX IF NOT EXISTS idx_devices_active ON devices(is_active);
CREATE INDEX IF NOT EXISTS idx_devices_last_seen ON devices(last_seen);

-- Insert sample device for testing
INSERT OR IGNORE INTO devices (
    device_id, 
    device_name, 
    device_type, 
    capabilities, 
    dashboard_url,
    last_seen,
    is_active,
    created_at,
    updated_at
) VALUES (
    'kindle_sample_device',
    'Sample Kindle Dashboard',
    'kindle',
    '["dashboard_display", "e_ink"]',
    '/',
    CURRENT_TIMESTAMP,
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);