-- Migration: Initial Schema
-- Version: 001

-- Schedules table
CREATE TABLE IF NOT EXISTS schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    time TEXT,
    label TEXT,
    audio_file TEXT
);

-- Audio files table
CREATE TABLE IF NOT EXISTS audio_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_name TEXT,
    display_name TEXT
);

-- Devices table
CREATE TABLE IF NOT EXISTS devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    ip_address TEXT,
    status TEXT,
    last_sync TEXT
);

-- Majors table
CREATE TABLE IF NOT EXISTS majors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT
);

-- Classes table
CREATE TABLE IF NOT EXISTS classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    major_id INTEGER,
    wa_group_id TEXT DEFAULT ''
);

-- Students table
CREATE TABLE IF NOT EXISTS students (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rfid_uid TEXT UNIQUE,
    nis TEXT,
    name TEXT,
    parent_phone TEXT,
    parent_name TEXT DEFAULT '',
    class_id INTEGER,
    photo TEXT
);

-- Staff table
CREATE TABLE IF NOT EXISTS staff (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rfid_uid TEXT UNIQUE,
    nip TEXT,
    name TEXT,
    phone TEXT,
    role TEXT
);

-- Attendance settings table
CREATE TABLE IF NOT EXISTS attendance_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    setting_key TEXT UNIQUE,
    setting_value TEXT
);

-- Attendance logs table
CREATE TABLE IF NOT EXISTS attendance_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rfid_uid TEXT,
    user_name TEXT,
    user_type TEXT,
    status TEXT,
    method TEXT DEFAULT 'RFID',
    timestamp DATETIME,
    date DATE
);

-- Holidays table
CREATE TABLE IF NOT EXISTS holidays (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- School settings table
CREATE TABLE IF NOT EXISTS school_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    setting_key TEXT UNIQUE NOT NULL,
    setting_value TEXT NOT NULL
);

-- WhatsApp logs table
CREATE TABLE IF NOT EXISTS whatsapp_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target TEXT,
    message TEXT,
    status TEXT,
    response TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Running texts table
CREATE TABLE IF NOT EXISTS running_texts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT,
    is_active BOOLEAN DEFAULT 1
);

-- Signage media table
CREATE TABLE IF NOT EXISTS signage_media (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT,
    file_type TEXT,
    duration INTEGER DEFAULT 10,
    is_active BOOLEAN DEFAULT 1
);

-- Prayer logs table
CREATE TABLE IF NOT EXISTS prayer_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rfid_uid TEXT,
    name TEXT,
    class_name TEXT,
    prayer_type TEXT,
    timestamp DATETIME,
    date DATE,
    status TEXT DEFAULT 'Hadir',
    recorded_by TEXT DEFAULT 'RFID'
);

-- Announcements table
CREATE TABLE IF NOT EXISTS announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    message TEXT,
    audio_file TEXT,
    scheduled_at DATETIME,
    played_at DATETIME,
    status TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Point rules table
CREATE TABLE IF NOT EXISTS point_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category TEXT,
    name TEXT,
    points INTEGER,
    description TEXT
);

-- Point rewards table
CREATE TABLE IF NOT EXISTS point_rewards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    points_cost INTEGER,
    stock INTEGER,
    description TEXT
);

-- Student points table
CREATE TABLE IF NOT EXISTS student_points (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER,
    rule_id INTEGER,
    reward_id INTEGER,
    points_change INTEGER,
    description TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    recorded_by TEXT
);

-- Operators table
CREATE TABLE IF NOT EXISTS operators (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    name TEXT NOT NULL,
    phone TEXT,
    photo TEXT,
    is_active BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
