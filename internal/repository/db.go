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
	cfg := config.GetDBConfig()
	dbDriver = cfg.Driver

	var dsn string
	var driverName string

	if cfg.Driver == "mysql" {
		driverName = "mysql"
		dsn = cfg.GetDSN()
		log.Printf("Connecting to MySQL: %s:%s/%s", cfg.Host, cfg.Port, cfg.Name)
	} else {
		driverName = "sqlite"
		dsn = cfg.Path
		log.Printf("Connecting to SQLite: %s", cfg.Path)
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
	SeedPointRules(db)

	return db
}

func SeedPointRules(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM point_rules").Scan(&count)
	if count > 0 {
		return
	}

	rules := []struct {
		code, category, name, description, tier string
		points int
	}{
		// A.1.1 Ibadah Shalat
		{"A.1.1.1", "A.1 - Keagamaan", "Shalat Dhuhur/Ashar insidental (diawasi)", "Melaksanakan shalat fardhu Dhuhur/Ashar di sekolah (insidental/diawasi)", "Dasar", 5},
		{"A.1.1.2", "A.1 - Keagamaan", "Shalat Dhuhur & Ashar rutin di sekolah", "Melaksanakan shalat fardhu Dhuhur & Ashar rutin di sekolah", "Pembiasaan", 10},
		{"A.1.1.3", "A.1 - Keagamaan", "Shalat 5 waktu (laporan mandiri)", "Melaksanakan shalat fardhu 5 waktu (laporan mandiri/buku mutaba'ah)", "Konsistensi", 15},
		{"A.1.1.4", "A.1 - Keagamaan", "Shalat 5 waktu tepat waktu", "Melaksanakan shalat fardhu 5 waktu selalu tepat waktu", "Disiplin", 20},
		{"A.1.1.5", "A.1 - Keagamaan", "Shalat 5 waktu berjamaah di masjid", "Melaksanakan shalat fardhu 5 waktu secara berjamaah di masjid/mushola", "Komunitas", 30},
		{"A.1.1.6", "A.1 - Keagamaan", "Shalat Dhuha insidental", "Melaksanakan Shalat Dhuha (insidental di sekolah)", "Dasar Sunnah", 10},
		{"A.1.1.7", "A.1 - Keagamaan", "Shalat Dhuha rutin", "Melaksanakan Shalat Dhuha rutin setiap hari sebelum KBM", "Istiqomah", 25},
		{"A.1.1.8", "A.1 - Keagamaan", "Shalat Tahajud", "Melaksanakan Shalat Tahajud (minimal 1-2 kali seminggu)", "Kesadaran", 30},
		{"A.1.1.9", "A.1 - Keagamaan", "Shalat Rawatib", "Melaksanakan Shalat Rawatib (Qobliyah/Ba'diyah) menyertai shalat wajib", "Pendalaman", 35},
		{"A.1.1.10", "A.1 - Keagamaan", "Muadzin shalat berjamaah", "Menjadi Muadzin untuk shalat berjamaah di sekolah", "Keberanian", 40},
		{"A.1.1.11", "A.1 - Keagamaan", "Imam shalat fardhu", "Menjadi Imam shalat fardhu di lingkungan sekolah/keluarga", "Kepemimpinan", 50},

		// A.1.2 Interaksi dengan Al-Qur'an
		{"A.1.2.1", "A.1 - Keagamaan", "Bisa membaca Al-Qur'an (Iqro)", "Mampu membaca Al-Qur'an (lulus Iqro/dasar)", "Dasar", 10},
		{"A.1.2.2", "A.1 - Keagamaan", "Baca Al-Qur'an tartil", "Membaca Al-Qur'an dengan tartil dan tajwid yang benar", "Kompetensi", 20},
		{"A.1.2.3", "A.1 - Keagamaan", "Tilawah 1 lembar/hari", "Rutin tilawah Al-Qur'an minimal 1 lembar per hari", "Pembiasaan", 25},
		{"A.1.2.4", "A.1 - Keagamaan", "One Day One Juz (ODOJ)", "Rutin tilawah Al-Qur'an One Day One Juz (ODOJ)", "Istiqomah", 40},
		{"A.1.2.5", "A.1 - Keagamaan", "Hafal surat pendek", "Hafal surat-surat pendek (Juz 30/Amma) minimal 10 surat", "Dasar Hafalan", 20},
		{"A.1.2.6", "A.1 - Keagamaan", "Hafal penuh Juz 30", "Hafal penuh Juz 30 (Juz Amma)", "Menengah", 40},
		{"A.1.2.7", "A.1 - Keagamaan", "Hafal 1-3 Juz di luar Juz 30", "Hafal 1-3 Juz Al-Qur'an (di luar Juz 30)", "Prestasi", 60},
		{"A.1.2.8", "A.1 - Keagamaan", "Hafal 5 Juz atau lebih", "Hafal 5 Juz Al-Qur'an atau lebih", "Visi Masa Depan", 80},
		{"A.1.2.9", "A.1 - Keagamaan", "Tutor Al-Qur'an sebaya", "Menjadi pembimbing/tutor sebaya membaca Al-Qur'an di sekolah", "Kepemimpinan", 50},

		// A.1.3 Puasa, Syiar, Kepemimpinan
		{"A.1.3.1", "A.1 - Keagamaan", "Puasa Ramadhan penuh", "Melaksanakan Puasa Ramadhan sebulan penuh", "Wajib", 20},
		{"A.1.3.2", "A.1 - Keagamaan", "Puasa Sunnah", "Melaksanakan Puasa Sunnah (Senin-Kamis atau Ayyamul Bidh)", "Kesadaran", 35},
		{"A.1.3.3", "A.1 - Keagamaan", "Ikut kajian rohis rutin", "Mengikuti kajian keislaman/rohis rutin di sekolah", "Partisipasi", 15},
		{"A.1.3.4", "A.1 - Keagamaan", "Kajian di luar sekolah", "Mengikuti kajian keislaman di luar sekolah (masjid raya/majelis taklim)", "Proaktif", 25},
		{"A.1.3.5", "A.1 - Keagamaan", "Panitia PHBI", "Aktif sebagai panitia Peringatan Hari Besar Islam (PHBI) di sekolah", "Tangkas/Sosial", 40},
		{"A.1.3.6", "A.1 - Keagamaan", "Kultum di sekolah", "Memberikan Kultum (Kuliah Tujuh Menit) di sekolah", "Komunikasi", 50},
		{"A.1.3.7", "A.1 - Keagamaan", "Ketua DKM/ROHIS", "Menjadi ketua DKM Sekolah / Ketua Rohis", "Kepemimpinan", 70},
		{"A.1.3.8", "A.1 - Keagamaan", "Lomba MTQ/Ceramah tingkat Kota", "Mewakili sekolah lomba MTQ/Ceramah/Nasyid tingkat Kota", "Kompetitif", 60},
		{"A.1.3.9", "A.1 - Keagamaan", "Juara MTQ/Ceramah Provinsi/Nasional", "Juara MTQ/Ceramah/Kaligrafi tingkat Provinsi/Nasional", "Visi Nasional", 100},

		// A.2 Sosial & Nasionalisme
		{"A.2.1", "A.2 - Sosial", "Upacara Bendera tepat waktu", "Hadir tepat waktu dan tertib mengikuti Upacara Bendera", "Nasionalisme Dasar", 10},
		{"A.2.2", "A.2 - Sosial", "Piket kelas bertanggung jawab", "Melaksanakan piket kelas dengan penuh tanggung jawab tanpa ditegur", "Disiplin", 15},
		{"A.2.3", "A.2 - Sosial", "Menjenguk teman/guru sakit", "Menjenguk/membantu teman atau guru yang sedang sakit/musibah", "Empati/Sosial", 20},
		{"A.2.4", "A.2 - Sosial", "Jaga fasilitas sekolah", "Aktif menjaga fasilitas sekolah (tidak membuang sampah sembarangan)", "Karakter", 25},
		{"A.2.5", "A.2 - Sosial", "Pettugas Upacara Bendera", "Menjadi petugas Upacara Bendera (Pengibar, Pembaca UUD, Dirijen)", "Keberanian", 35},
		{"A.2.6", "A.2 - Sosial", "Inisiator donasi/bantuan", "Menginisiasi penggalangan dana/donasi untuk bencana/sosial", "Inisiatif/Pemimpin", 50},
		{"A.2.7", "A.2 - Sosial", "Pengurus OSIS/MPK", "Menjadi Pengurus OSIS, MPK, atau Komandan Paskibra/PMR", "Kepemimpinan", 60},
		{"A.2.8", "A.2 - Sosial", "Paskibraka/Duta Pemuda Kota", "Mewakili sekolah dalam ajang Paskibraka/Duta Pemuda tingkat Kota", "Prestasi Kota", 75},
		{"A.2.9", "A.2 - Sosial", "Paskibraka/Duta Pemuda Provinsi/Nasional", "Terpilih sebagai Paskibraka/Duta Pemuda tingkat Provinsi/Nasional", "Visi Nasional", 100},

		// A.3 Kewirausahaan
		{"A.3.1", "A.3 - Kewirausahaan", "Baca literatur bisnis", "Aktif berdiskusi dan membaca literatur/buku bisnis (di luar modul wajib)", "Minat Belajar", 10},
		{"A.3.2", "A.3 - Kewirausahaan", "Susun Business Model Canvas", "Mampu menyusun Business Model Canvas (BMC) untuk ide usaha", "Cerdas/Kreatif", 25},
		{"A.3.3", "A.3 - Kewirausahaan", "Berjualan di sekolah", "Berani berjualan/menawarkan produk secara langsung di lingkungan sekolah", "Tangkas/Mental", 35},
		{"A.3.4", "A.3 - Kewirausahaan", "Gunakan media sosial untuk jualan", "Memanfaatkan media sosial/e-commerce secara aktif untuk berjualan", "Inovatif", 40},
		{"A.3.5", "A.3 - Kewirausahaan", "hasilkan omzet konsisten", "Menghasilkan omzet dari usaha mandiri secara konsisten (laporan bulanan)", "Praktisi Bisnis", 50},
		{"A.3.6", "A.3 - Kewirausahaan", "Kolaborasi bisnis proyek", "Menginisiasi kolaborasi bisnis/proyek dengan teman lintas kelas/jurusan", "Kolaboratif", 60},
		{"A.3.7", "A.3 - Kewirausahaan", "Produk diliput media", "Produk/jasanya diliput media atau diakui oleh pihak eksternal/industri", "Validasi Pasar", 70},
		{"A.3.8", "A.3 - Kewirausahaan", "Lomba Business Plan Kota", "Mewakili sekolah lomba Business Plan/Inovasi Bisnis tingkat Kota", "Kompetitif", 75},
		{"A.3.9", "A.3 - Kewirausahaan", "Juara Bisnis/Nasional atau punya IUMK", "Juara 1,2,3 kompetisi wirausaha tingkat Nasional / Memiliki IUMK", "Visi Masa Depan", 100},

		// A.4 Akademik & Bahasa
		{"A.4.1", "A.4 - Akademik", "Kehadiran kelas 100%", "Tingkat kehadiran kelas 100% dalam satu semester", "Disiplin", 20},
		{"A.4.2", "A.4 - Akademik", "Aktif di kelas", "Aktif bertanya, menjawab, dan berdiskusi di kelas", "Tangkas/Cerdas", 25},
		{"A.4.3", "A.4 - Akademik", "Peringkat 10 besar", "Masuk peringkat 10 besar paralel di angkatan", "Konsistensi", 40},
		{"A.4.4", "A.4 - Akademik", "Presentasi Bahasa Inggris", "Berani presentasi menggunakan Bahasa Inggris/Asing di depan kelas", "Visi Global", 45},
		{"A.4.5", "A.4 - Akademik", "Sertifikasi keahlian", "Lulus sertifikasi keahlian komputer/akuntansi/bahasa (TOEFL/TOEIC)", "Kompetensi Ekstra", 60},
		{"A.4.6", "A.4 - Akademik", "Juara Kelas/Umum", "Juara 1 Kelas atau Juara Umum Angkatan", "Prestasi Puncak", 70},
		{"A.4.7", "A.4 - Akademik", "Lomba Akademik Kota", "Mengikuti Olympiade Siswa/LKS tingkat Kota/Kabupaten", "Kompetitif", 65},
		{"A.4.8", "A.4 - Akademik", "Juara LKS/Olympiade Nasional", "Meraih Medali/Juara LKS/Olympiade Akademik tingkat Nasional", "Visi Nasional", 100},

		// A.5 Kesehatan, Bakat, Seni & Olahraga
		{"A.5.1", "A.5 - Ekstrakurikuler", "Senam/olahraga antusias", "Rutin mengikuti senam/olahraga bersama dengan antusias", "Sehat/Bahagia", 10},
		{"A.5.2", "A.5 - Ekstrakurikuler", "Sikap 5S (Senyum, Salam, Sopan, Santun, Steril)", "Menunjukkan sikap ceria, ramah (5S) dan membawa energi positif di sekolah", "Bahagia/Karakter", 20},
		{"A.5.3", "A.5 - Ekstrakurikuler", "Aktif ekskul", "Terdaftar dan aktif minimal dalam 1 kegiatan Ekstrakurikuler", "Eksplorasi Diri", 25},
		{"A.5.4", "A.5 - Ekstrakurikuler", "Ciptakan karya seni", "Menciptakan karya seni (lukisan, puisi, musik, desain) yang dipublikasikan", "Kreatif/Inovatif", 40},
		{"A.5.5", "A.5 - Ekstrakurikuler", "Kelola akun publikasi", "Mengelola akun publikasi/jurnalistik/kreatif sekolah dengan baik", "Inovatif", 50},
		{"A.5.6", "A.5 - Ekstrakurikuler", "Kapten tim olahraga/seni", "Menjadi anggota inti/kapten tim olahraga atau kesenian sekolah", "Kepemimpinan Tim", 60},
		{"A.5.7", "A.5 - Ekstrakurikuler", "Juara O2SN/FLS2N Kota", "Juara kompetisi olahraga/seni (O2SN/FLS2N) tingkat Kota", "Kompetitif", 75},
		{"A.5.8", "A.5 - Ekstrakurikuler", "Juara O2SN/FLS2N Nasional", "Juara kompetisi olahraga/seni tingkat Provinsi/Nasional", "Visi Nasional", 100},
	}

	for _, r := range rules {
		if config.IsMySQL() {
			db.Exec(`INSERT IGNORE INTO point_rules (code, category, name, description, points, tier) VALUES (?, ?, ?, ?, ?, ?)`,
				r.code, r.category, r.name, r.description, r.points, r.tier)
		} else {
			db.Exec(`INSERT OR IGNORE INTO point_rules (code, category, name, description, points, tier) VALUES (?, ?, ?, ?, ?, ?)`,
				r.code, r.category, r.name, r.description, r.points, r.tier)
		}
	}
}

func RunMigrations(db *sql.DB) {
	if config.IsMySQL() {
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
		{"students", "nis_siswa", "ALTER TABLE students ADD COLUMN nis_siswa TEXT DEFAULT ''"},
		{"attendance_settings", "point_claim_enabled", "ALTER TABLE attendance_settings ADD COLUMN point_claim_enabled TEXT DEFAULT 'true'"},
	}

	// Create tables if not exist
	db.Exec(`
		CREATE TABLE IF NOT EXISTS point_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL,
			category TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			points INTEGER NOT NULL,
			tier TEXT NOT NULL,
			is_active INTEGER DEFAULT 1
		)`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS point_claims (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			rule_id INTEGER NOT NULL,
			description TEXT,
			evidence TEXT,
			status TEXT DEFAULT 'pending',
			submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			reviewed_at DATETIME,
			reviewed_by INTEGER,
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(rule_id) REFERENCES point_rules(id)
		)`)

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
		{"students", "nis_siswa", "ALTER TABLE students ADD COLUMN nis_siswa VARCHAR(50) DEFAULT ''"},
		{"attendance_settings", "point_claim_enabled", "ALTER TABLE attendance_settings ADD COLUMN point_claim_enabled VARCHAR(10) DEFAULT 'true'"},
	}

	// Create tables if not exist for MySQL
	db.Exec(`
		CREATE TABLE IF NOT EXISTS point_rules (
			id INT PRIMARY KEY AUTO_INCREMENT,
			code VARCHAR(20) NOT NULL,
			category VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			points INT NOT NULL,
			tier VARCHAR(50) NOT NULL,
			is_active INT DEFAULT 1
		)`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS point_claims (
			id INT PRIMARY KEY AUTO_INCREMENT,
			student_id INT NOT NULL,
			rule_id INT NOT NULL,
			description TEXT,
			evidence TEXT,
			status VARCHAR(20) DEFAULT 'pending',
			submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			reviewed_at DATETIME,
			reviewed_by INT,
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(rule_id) REFERENCES point_rules(id)
		)`)

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
		if config.IsMySQL() {
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
