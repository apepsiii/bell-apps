package repository

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"

	"belsekolah/internal/config"
)

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite", config.DBPath)
	if err != nil {
		log.Fatal("Gagal membuka database:", err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			time TEXT,
			label TEXT,
			audio_file TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS audio_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name TEXT,
			display_name TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			ip_address TEXT,
			status TEXT,
			last_sync TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS majors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS classes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			major_id INTEGER,
			wa_group_id TEXT DEFAULT ''
		);`,
		`CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rfid_uid TEXT UNIQUE,
			nis TEXT,
			name TEXT,
			parent_phone TEXT,
			parent_name TEXT DEFAULT '',
			class_id INTEGER,
			photo TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS staff (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rfid_uid TEXT UNIQUE,
			nip TEXT,
			name TEXT,
			phone TEXT,
			role TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS attendance_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setting_key TEXT UNIQUE,
			setting_value TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS holidays (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS school_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setting_key TEXT UNIQUE NOT NULL,
			setting_value TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS whatsapp_logs (id INTEGER PRIMARY KEY AUTOINCREMENT, target TEXT, message TEXT, status TEXT, response TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP);`,
		`INSERT OR IGNORE INTO school_settings (setting_key, setting_value) VALUES ('work_days', '1,2,3,4,5');`,
		`CREATE TABLE IF NOT EXISTS attendance_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rfid_uid TEXT,
			user_name TEXT,
			user_type TEXT,
			status TEXT,
			method TEXT DEFAULT 'RFID',
			timestamp DATETIME,
			date DATE
		);`,
		`CREATE TABLE IF NOT EXISTS running_texts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT,
			is_active BOOLEAN DEFAULT 1
		);`,
		`CREATE TABLE IF NOT EXISTS signage_media (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT,
			file_type TEXT,
			duration INTEGER DEFAULT 10,
			is_active BOOLEAN DEFAULT 1
		);`,
		`CREATE TABLE IF NOT EXISTS prayer_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rfid_uid TEXT,
			name TEXT,
			class_name TEXT,
			prayer_type TEXT,
			timestamp DATETIME,
			date DATE,
			status TEXT DEFAULT 'Hadir',
			recorded_by TEXT DEFAULT 'RFID'
		);`,
		`CREATE TABLE IF NOT EXISTS announcements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			message TEXT,
			audio_file TEXT,
			scheduled_at DATETIME,
			played_at DATETIME,
			status TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS point_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category TEXT,
			name TEXT,
			points INTEGER,
			description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS point_rewards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			points_cost INTEGER,
			stock INTEGER,
			description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS student_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER,
			rule_id INTEGER,
			reward_id INTEGER,
			points_change INTEGER,
			description TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			recorded_by TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS operators (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			phone TEXT,
			photo TEXT,
			is_active BOOLEAN DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`INSERT OR IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('dzuhur_start', '11:30');`,
		`INSERT OR IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('dzuhur_end', '13:00');`,
		`INSERT OR IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('ashar_start', '15:00');`,
		`INSERT OR IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('ashar_end', '16:00');`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Fatal("Gagal migrasi tabel:", err)
		}
	}

	RunMigrations(db)
	SeedDefaultData(db)

	return db
}

func RunMigrations(db *sql.DB) {
	migrations := []struct {
		table  string
		column string
		alter  string
	}{
		{"attendance_logs", "method", "ALTER TABLE attendance_logs ADD COLUMN method TEXT DEFAULT 'RFID'"},
		{"classes", "wa_group_id", "ALTER TABLE classes ADD COLUMN wa_group_id TEXT DEFAULT ''"},
		{"prayer_logs", "status", "ALTER TABLE prayer_logs ADD COLUMN status TEXT DEFAULT 'Hadir'"},
		{"students", "parent_name", "ALTER TABLE students ADD COLUMN parent_name TEXT DEFAULT ''"},
		{"prayer_logs", "recorded_by", "ALTER TABLE prayer_logs ADD COLUMN recorded_by TEXT DEFAULT 'RFID'"},
		{"students", "birthday", "ALTER TABLE students ADD COLUMN birthday TEXT DEFAULT ''"},
	}

	for _, m := range migrations {
		var colCount int
		db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('" + m.table + "') WHERE name='" + m.column + "'").Scan(&colCount)
		if colCount == 0 {
			db.Exec(m.alter)
		}
	}
}

func SeedDefaultData(db *sql.DB) {
	var countSettings int
	db.QueryRow("SELECT COUNT(*) FROM attendance_settings").Scan(&countSettings)
	if countSettings == 0 {
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "arrival_start", "06:00")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "arrival_end", "07:15")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "departure_start", "15:30")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "departure_end", "17:00")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "onesender_api_url", "https://onesender.my.id/api/v1/messages")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "onesender_api_token", "")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_template_in", "Halo, Ananda {name} telah hadir di sekolah pada pukul {time}. Status: {status}.")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_template_late", "Halo, Ananda {name} terlambat hadir di sekolah pada pukul {time}.")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_template_out", "Halo, Ananda {name} telah pulang sekolah pada pukul {time}.")

		staffIn := "✅ KONFIRMASI KEDATANGAN GURU/STAF\n\nYth. Bapak/Ibu {teacher_name},\n\nPresensi {type} Anda pada hari {date} telah berhasil dicatat sistem.\n\n🕒 Pukul: {time} WIB\n\nSelamat bertugas dan semoga hari Anda menyenangkan!\n\n— Sistem Presensi Sekolah —"
		staffOut := "✅ KONFIRMASI KEPULANGAN GURU/STAF\n\nYth. Bapak/Ibu {teacher_name},\n\nPresensi {type} Anda pada hari {date} telah berhasil dicatat sistem.\n\n🕒 Pukul: {time} WIB\n\nTerima kasih atas dedikasi hari ini. Selamat beristirahat.\n\n— Sistem Presensi Sekolah —"

		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_template_staff_in", staffIn)
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_template_staff_out", staffOut)
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_image_link", "https://via.placeholder.com/150")
	}
}
