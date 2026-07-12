package repository

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"

	"belsekolah/internal/config"
)

var dbDriver string

func InitDB() *sql.DB {
	cfg := config.GetDatabaseConfig()
	dbDriver = cfg.Driver

	var dsn string
	var driverName string

	if cfg.Driver == "mysql" {
		driverName = "mysql"
		dsn = cfg.GetDSN()
		log.Printf("Connecting to MySQL: %s:%s/%s", cfg.Host, cfg.Port, cfg.Database)
	} else {
		driverName = "sqlite"
		dsn = cfg.DBPath
		log.Printf("Connecting to SQLite: %s", cfg.DBPath)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		log.Fatal("Gagal membuka database:", err)
	}

	if cfg.Driver == "mysql" {
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Gagal ping database:", err)
	}

	RunMigrations(db)
	SeedDefaultData(db)

	return db
}

func IsMySQL() bool {
	return dbDriver == "mysql"
}

func RunMigrations(db *sql.DB) {
	if IsMySQL() {
		runMySQLMigrations(db)
	} else {
		runSQLiteMigrations(db)
	}
}

func runSQLiteMigrations(db *sql.DB) {
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
		{"students", "status", "ALTER TABLE students ADD COLUMN status TEXT DEFAULT 'active'"},
	}

	for _, m := range migrations {
		var colCount int
		db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('" + m.table + "') WHERE name='" + m.column + "'").Scan(&colCount)
		if colCount == 0 {
			db.Exec(m.alter)
		}
	}
}

func runMySQLMigrations(db *sql.DB) {
	migrations := []struct {
		table  string
		column string
		alter  string
	}{
		{"attendance_logs", "method", "ALTER TABLE attendance_logs ADD COLUMN method VARCHAR(50) DEFAULT 'RFID'"},
		{"classes", "wa_group_id", "ALTER TABLE classes ADD COLUMN wa_group_id VARCHAR(255) DEFAULT ''"},
		{"prayer_logs", "status", "ALTER TABLE prayer_logs ADD COLUMN status VARCHAR(50) DEFAULT 'Hadir'"},
		{"students", "parent_name", "ALTER TABLE students ADD COLUMN parent_name VARCHAR(255) DEFAULT ''"},
		{"prayer_logs", "recorded_by", "ALTER TABLE prayer_logs ADD COLUMN recorded_by VARCHAR(50) DEFAULT 'RFID'"},
		{"students", "birthday", "ALTER TABLE students ADD COLUMN birthday VARCHAR(20) DEFAULT ''"},
		{"students", "status", "ALTER TABLE students ADD COLUMN status VARCHAR(20) DEFAULT 'active'"},
	}

	for _, m := range migrations {
		var colCount int
		db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?", m.table, m.column).Scan(&colCount)
		if colCount == 0 {
			db.Exec(m.alter)
		}
	}
}

func SeedDefaultData(db *sql.DB) {
	var countSettings int
	db.QueryRow("SELECT COUNT(*) FROM attendance_settings").Scan(&countSettings)
	if countSettings == 0 {
		if IsMySQL() {
			seedMySQLData(db)
		} else {
			seedSQLiteData(db)
		}
	}
}

func seedSQLiteData(db *sql.DB) {
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
	db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "dzuhur_start", "11:30")
	db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "dzuhur_end", "13:00")
	db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "ashar_start", "15:00")
	db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "ashar_end", "16:00")
	db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "birthday_enabled", "false")
	db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "birthday_time", "08:00")
	db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_template_birthday", "🎂🎉 Selamat Ulang Tahun, {name}! 🎉🎂\n\nSemoga tahun ini dipenuhi kebahagiaan!")
	db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_image_birthday", "https://via.placeholder.com/300x300?text=Happy+Birthday")
	db.Exec("INSERT INTO school_settings VALUES (?, ?)", "work_days", "1,2,3,4,5")
}

func seedMySQLData(db *sql.DB) {
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "arrival_start", "06:00")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "arrival_end", "07:15")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "departure_start", "15:30")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "departure_end", "17:00")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "onesender_api_url", "https://onesender.my.id/api/v1/messages")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "onesender_api_token", "")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_in", "Halo, Ananda {name} telah hadir di sekolah pada pukul {time}. Status: {status}.")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_late", "Halo, Ananda {name} terlambat hadir di sekolah pada pukul {time}.")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_out", "Halo, Ananda {name} telah pulang sekolah pada pukul {time}.")

	staffIn := "✅ KONFIRMASI KEDATANGAN GURU/STAF\n\nYth. Bapak/Ibu {teacher_name},\n\nPresensi {type} Anda pada hari {date} telah berhasil dicatat sistem.\n\n🕒 Pukul: {time} WIB\n\nSelamat bertugas dan semoga hari Anda menyenangkan!\n\n— Sistem Presensi Sekolah —"
	staffOut := "✅ KONFIRMASI KEPULANGAN GURU/STAF\n\nYth. Bapak/Ibu {teacher_name},\n\nPresensi {type} Anda pada hari {date} telah berhasil dicatat sistem.\n\n🕒 Pukul: {time} WIB\n\nTerima kasih atas dedikasi hari ini. Selamat beristirahat.\n\n— Sistem Presensi Sekolah —"

	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_staff_in", staffIn)
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_staff_out", staffOut)
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_image_link", "https://via.placeholder.com/150")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "dzuhur_start", "11:30")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "dzuhur_end", "13:00")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "ashar_start", "15:00")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "ashar_end", "16:00")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "birthday_enabled", "false")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "birthday_time", "08:00")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_birthday", "🎂🎉 Selamat Ulang Tahun, {name}! 🎉🎂\n\nSemoga tahun ini dipenuhi kebahagiaan!")
	db.Exec("INSERT IGNORE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_image_birthday", "https://via.placeholder.com/300x300?text=Happy+Birthday")
	db.Exec("INSERT IGNORE INTO school_settings (setting_key, setting_value) VALUES (?, ?)", "work_days", "1,2,3,4,5")
}
