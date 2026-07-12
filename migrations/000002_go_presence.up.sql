-- Migration: Go Presence Schema
-- Version: 000002

-- Schedules table
CREATE TABLE IF NOT EXISTS schedules (
    id INT PRIMARY KEY AUTO_INCREMENT,
    time TEXT,
    label TEXT,
    audio_file TEXT
);

-- Audio files table
CREATE TABLE IF NOT EXISTS audio_files (
    id INT PRIMARY KEY AUTO_INCREMENT,
    file_name TEXT,
    display_name TEXT
);

-- Devices table
CREATE TABLE IF NOT EXISTS devices (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name TEXT,
    ip_address TEXT,
    status TEXT,
    last_sync TEXT
);

-- Majors table
CREATE TABLE IF NOT EXISTS majors (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name TEXT
);

-- Classes table
CREATE TABLE IF NOT EXISTS classes (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name TEXT,
    major_id INT,
    wa_group_id TEXT
);

-- Students table
CREATE TABLE IF NOT EXISTS students (
    id INT PRIMARY KEY AUTO_INCREMENT,
    rfid_uid VARCHAR(255) UNIQUE,
    nis VARCHAR(255),
    name VARCHAR(255),
    parent_phone VARCHAR(50),
    parent_name VARCHAR(255) DEFAULT '',
    class_id INT,
    photo VARCHAR(255),
    birthday VARCHAR(20) DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Staff table
CREATE TABLE IF NOT EXISTS staff (
    id INT PRIMARY KEY AUTO_INCREMENT,
    rfid_uid VARCHAR(255) UNIQUE,
    nip VARCHAR(255),
    name VARCHAR(255),
    phone VARCHAR(50),
    role VARCHAR(100)
);

-- Attendance settings table
CREATE TABLE IF NOT EXISTS attendance_settings (
    id INT PRIMARY KEY AUTO_INCREMENT,
    setting_key VARCHAR(100) UNIQUE,
    setting_value TEXT
);

-- Attendance logs table
CREATE TABLE IF NOT EXISTS attendance_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    rfid_uid VARCHAR(255),
    user_name VARCHAR(255),
    user_type VARCHAR(50),
    status VARCHAR(50),
    method VARCHAR(50) DEFAULT 'RFID',
    timestamp DATETIME,
    date DATE,
    INDEX idx_rfid (rfid_uid),
    INDEX idx_date (date)
);

-- Holidays table
CREATE TABLE IF NOT EXISTS holidays (
    id INT PRIMARY KEY AUTO_INCREMENT,
    date DATE NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- School settings table
CREATE TABLE IF NOT EXISTS school_settings (
    id INT PRIMARY KEY AUTO_INCREMENT,
    setting_key VARCHAR(100) UNIQUE NOT NULL,
    setting_value TEXT NOT NULL
);

-- WhatsApp logs table
CREATE TABLE IF NOT EXISTS whatsapp_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    target VARCHAR(100),
    message TEXT,
    status VARCHAR(50),
    response TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Running texts table
CREATE TABLE IF NOT EXISTS running_texts (
    id INT PRIMARY KEY AUTO_INCREMENT,
    content TEXT,
    is_active BOOLEAN DEFAULT 1
);

-- Signage media table
CREATE TABLE IF NOT EXISTS signage_media (
    id INT PRIMARY KEY AUTO_INCREMENT,
    filename TEXT,
    file_type TEXT,
    duration INT DEFAULT 10,
    is_active BOOLEAN DEFAULT 1
);

-- Prayer logs table
CREATE TABLE IF NOT EXISTS prayer_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    rfid_uid VARCHAR(255),
    name VARCHAR(255),
    class_name VARCHAR(255),
    prayer_type VARCHAR(50),
    timestamp DATETIME,
    date DATE,
    status VARCHAR(50) DEFAULT 'Hadir',
    recorded_by VARCHAR(50) DEFAULT 'RFID',
    INDEX idx_rfid (rfid_uid),
    INDEX idx_date (date)
);

-- Announcements table
CREATE TABLE IF NOT EXISTS announcements (
    id INT PRIMARY KEY AUTO_INCREMENT,
    title TEXT,
    message TEXT,
    audio_file TEXT,
    scheduled_at DATETIME,
    played_at DATETIME,
    status VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Point rules table
CREATE TABLE IF NOT EXISTS point_rules (
    id INT PRIMARY KEY AUTO_INCREMENT,
    category VARCHAR(100),
    name VARCHAR(255),
    points INT,
    description TEXT
);

-- Point rewards table
CREATE TABLE IF NOT EXISTS point_rewards (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255),
    points_cost INT,
    stock INT,
    description TEXT
);

-- Student points table
CREATE TABLE IF NOT EXISTS student_points (
    id INT PRIMARY KEY AUTO_INCREMENT,
    student_id INT,
    rule_id INT,
    reward_id INT,
    points_change INT,
    description TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    recorded_by VARCHAR(100),
    INDEX idx_student (student_id)
);

-- Operators table
CREATE TABLE IF NOT EXISTS operators (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    photo VARCHAR(255),
    is_active BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Face encodings table (for face recognition)
CREATE TABLE IF NOT EXISTS student_faces (
    id INT PRIMARY KEY AUTO_INCREMENT,
    student_id INT NOT NULL,
    encoding MEDIUMTEXT NOT NULL,
    image_path VARCHAR(500),
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    UNIQUE KEY unique_student_face (student_id)
);

-- Insert default school settings
INSERT IGNORE INTO school_settings (setting_key, setting_value) VALUES ('work_days', '1,2,3,4,5');

-- Insert default attendance settings
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('arrival_start', '06:00');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('arrival_end', '07:15');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('departure_start', '15:30');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('departure_end', '17:00');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('onesender_api_url', 'https://onesender.my.id/api/v1/messages');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('onesender_api_token', '');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('wa_template_in', 'Halo, Ananda {name} telah hadir di sekolah pada pukul {time}. Status: {status}.');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('wa_template_late', 'Halo, Ananda {name} terlambat hadir di sekolah pada pukul {time}.');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('wa_template_out', 'Halo, Ananda {name} telah pulang sekolah pada pukul {time}.');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('wa_image_link', 'https://via.placeholder.com/150');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('dzuhur_start', '11:30');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('dzuhur_end', '13:00');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('ashar_start', '15:00');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('ashar_end', '16:00');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('birthday_enabled', 'false');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('birthday_time', '08:00');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('wa_template_birthday', '🎂🎉 Selamat Ulang Tahun, {name}! 🎉🎂\n\nSemoga tahun ini dipenuhi kebahagiaan!');
INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('wa_image_birthday', 'https://via.placeholder.com/300x300?text=Happy+Birthday');
