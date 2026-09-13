package repository

import (
	"database/sql"
	"fmt"
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
	SeedEnglishBadges(db)
	SeedAttendanceInsights(db)
	SeedAssetData(db)
	SeedDualTrackPointRules(db)
	
	// Run audit trail migration
	if err := MigrationAuditTrail(db); err != nil {
		log.Printf("Warning: Audit trail migration failed: %v", err)
	} else {
		log.Println("✅ Audit trail migration completed successfully")
	}

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
		// point_rules table schema: (id, category, name, points, description).
		// The rule code/tier are kept in the struct for reference only and are
		// folded into the description so the seed still carries that context.
		fullDesc := r.description
		if r.code != "" || r.tier != "" {
			fullDesc = fmt.Sprintf("[%s | %s] %s", r.code, r.tier, r.description)
		}
		if config.IsMySQL() {
			db.Exec(`INSERT IGNORE INTO point_rules (category, name, points, description) VALUES (?, ?, ?, ?)`,
				r.category, r.name, r.points, fullDesc)
		} else {
			db.Exec(`INSERT OR IGNORE INTO point_rules (category, name, points, description) VALUES (?, ?, ?, ?)`,
				r.category, r.name, r.points, fullDesc)
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
	// === Core schema: create all tables IF NOT EXISTS ===
	// This runs on every startup and is idempotent. Schema matches the
	// production DB so a fresh deployment bootstraps correctly.
	schema := []string{
		`CREATE TABLE IF NOT EXISTS schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			time TEXT,
			label TEXT,
			audio_file TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS audio_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name TEXT,
			display_name TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS majors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS classes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			major_id INTEGER,
			wa_group_id TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rfid_uid TEXT UNIQUE,
			nis TEXT,
			name TEXT,
			parent_phone TEXT,
			class_id INTEGER,
			photo TEXT DEFAULT '',
			parent_name TEXT DEFAULT '',
			birthday TEXT DEFAULT '',
			status TEXT DEFAULT 'active',
			nis_siswa TEXT DEFAULT '',
			password TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS staff (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rfid_uid TEXT UNIQUE,
			nip TEXT,
			name TEXT,
			phone TEXT,
			role TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS attendance_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rfid_uid TEXT,
			user_name TEXT,
			user_type TEXT,
			status TEXT,
			timestamp DATETIME,
			date DATE,
			method TEXT DEFAULT 'RFID'
		)`,
		`CREATE TABLE IF NOT EXISTS attendance_settings (
			setting_key TEXT PRIMARY KEY,
			setting_value TEXT,
			point_claim_enabled TEXT DEFAULT 'true'
		)`,
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
		)`,
		`CREATE TABLE IF NOT EXISTS whatsapp_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			target TEXT,
			message TEXT,
			status TEXT,
			response TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS student_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER,
			rule_id INTEGER,
			reward_id INTEGER,
			points_change INTEGER,
			description TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			recorded_by TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			ip_address TEXT,
			status TEXT,
			last_sync TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS announcements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			message TEXT,
			audio_file TEXT,
			scheduled_at DATETIME,
			played_at DATETIME,
			status TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS holidays (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS school_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setting_key TEXT UNIQUE NOT NULL,
			setting_value TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS operators (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			phone TEXT,
			photo TEXT,
			is_active BOOLEAN DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS running_texts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT,
			is_active BOOLEAN DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS signage_media (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT,
			file_type TEXT,
			duration INTEGER DEFAULT 10,
			is_active BOOLEAN DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS point_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category TEXT,
			name TEXT,
			points INTEGER,
			description TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS point_claims (
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
		)`,
		`CREATE TABLE IF NOT EXISTS point_rewards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			points_cost INTEGER,
			stock INTEGER,
			description TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS attendance_insights (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category TEXT NOT NULL,
			message TEXT NOT NULL,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS english_quests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			description TEXT,
			quest_type TEXT NOT NULL DEFAULT 'written',
			topic TEXT DEFAULT '',
			vocabulary_words TEXT DEFAULT '',
			quiz_question TEXT DEFAULT '',
			quiz_choices TEXT DEFAULT '',
			quiz_answer TEXT DEFAULT '',
			xp_reward INTEGER DEFAULT 10,
			created_by TEXT DEFAULT 'admin',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS english_submissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			quest_id INTEGER NOT NULL,
			quest_type TEXT NOT NULL,
			content TEXT DEFAULT '',
			audio_file TEXT DEFAULT '',
			vocab_words TEXT DEFAULT '',
			quiz_answer TEXT DEFAULT '',
			xp_earned INTEGER DEFAULT 0,
			status TEXT DEFAULT 'pending',
			feedback TEXT DEFAULT '',
			submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			reviewed_at DATETIME,
			reviewed_by TEXT DEFAULT '',
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(quest_id) REFERENCES english_quests(id)
		)`,
		`CREATE TABLE IF NOT EXISTS english_streaks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL UNIQUE,
			current_streak INTEGER DEFAULT 0,
			longest_streak INTEGER DEFAULT 0,
			total_xp INTEGER DEFAULT 0,
			last_submit_date TEXT DEFAULT '',
			FOREIGN KEY(student_id) REFERENCES students(id)
		)`,
		`CREATE TABLE IF NOT EXISTS english_badges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			icon TEXT DEFAULT '🏅',
			condition_type TEXT NOT NULL,
			condition_value INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS english_student_badges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			badge_id INTEGER NOT NULL,
			earned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(student_id, badge_id),
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(badge_id) REFERENCES english_badges(id)
		)`,
		`CREATE TABLE IF NOT EXISTS ai_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setting_key TEXT UNIQUE NOT NULL,
			setting_value TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// === DUAL-TRACK POINT SYSTEM ===
		`CREATE TABLE IF NOT EXISTS achievement_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			category TEXT NOT NULL,
			category_code TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			points INTEGER NOT NULL DEFAULT 0,
			min_points INTEGER DEFAULT 0,
			max_points INTEGER DEFAULT 0,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS violation_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			category TEXT NOT NULL,
			category_code TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			points_1 INTEGER NOT NULL DEFAULT 0,
			points_2 INTEGER NOT NULL DEFAULT 0,
			points_3 INTEGER NOT NULL DEFAULT 0,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS student_achievement_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			rule_id INTEGER,
			points INTEGER NOT NULL DEFAULT 0,
			description TEXT,
			recorded_by TEXT DEFAULT 'Admin',
			academic_year TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(rule_id) REFERENCES achievement_rules(id)
		)`,
		`CREATE TABLE IF NOT EXISTS student_violation_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			rule_id INTEGER,
			occurrence INTEGER DEFAULT 1,
			points INTEGER NOT NULL DEFAULT 0,
			description TEXT,
			recorded_by TEXT DEFAULT 'Admin',
			academic_year TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(rule_id) REFERENCES violation_rules(id)
		)`,
		// === SARPRAS TABLES ===
		`CREATE TABLE IF NOT EXISTS asset_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_funding_sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_locations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			type TEXT,
			capacity INTEGER,
			pic_staff_id INTEGER,
			status TEXT DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS assets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			inventory_code TEXT NOT NULL UNIQUE,
			qr_code TEXT UNIQUE,
			category_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			specification TEXT,
			brand TEXT,
			serial_number TEXT,
			acquisition_year INTEGER,
			acquisition_date DATE,
			purchase_price REAL,
			funding_source_id INTEGER,
			location_id INTEGER,
			pic_staff_id INTEGER,
			condition TEXT DEFAULT 'good',
			quantity INTEGER DEFAULT 1,
			unit TEXT,
			notes TEXT,
			photo_url TEXT,
			is_borrowable INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME,
			created_by INTEGER,
			FOREIGN KEY (category_id) REFERENCES asset_categories(id),
			FOREIGN KEY (location_id) REFERENCES asset_locations(id),
			FOREIGN KEY (funding_source_id) REFERENCES asset_funding_sources(id)
		)`,
		`CREATE TABLE IF NOT EXISTS asset_borrowings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			asset_id INTEGER NOT NULL,
			borrower_type TEXT NOT NULL,
			borrower_id INTEGER NOT NULL,
			purpose TEXT NOT NULL,
			borrow_date DATETIME NOT NULL,
			due_date DATETIME NOT NULL,
			return_date DATETIME,
			status TEXT DEFAULT 'pending',
			approved_by INTEGER,
			approved_at DATETIME,
			rejection_reason TEXT,
			return_condition TEXT,
			return_notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME,
			FOREIGN KEY (asset_id) REFERENCES assets(id)
		)`,
		`CREATE TABLE IF NOT EXISTS asset_maintenance_tickets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ticket_number TEXT NOT NULL UNIQUE,
			asset_id INTEGER,
			location_id INTEGER,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			priority TEXT DEFAULT 'medium',
			status TEXT DEFAULT 'open',
			reported_by INTEGER NOT NULL,
			reported_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			assigned_to INTEGER,
			photo_urls TEXT,
			resolution_notes TEXT,
			resolved_at DATETIME,
			closed_at DATETIME,
			cost REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME,
			FOREIGN KEY (asset_id) REFERENCES assets(id),
			FOREIGN KEY (location_id) REFERENCES asset_locations(id)
		)`,
		`CREATE TABLE IF NOT EXISTS asset_movements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			asset_id INTEGER NOT NULL,
			from_location_id INTEGER,
			to_location_id INTEGER NOT NULL,
			moved_by INTEGER NOT NULL,
			moved_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			reason TEXT,
			notes TEXT,
			FOREIGN KEY (asset_id) REFERENCES assets(id),
			FOREIGN KEY (from_location_id) REFERENCES asset_locations(id),
			FOREIGN KEY (to_location_id) REFERENCES asset_locations(id)
		)`,
		`CREATE TABLE IF NOT EXISTS asset_disposals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			asset_id INTEGER NOT NULL,
			disposal_date DATE NOT NULL,
			reason TEXT NOT NULL,
			description TEXT,
			book_value REAL,
			disposal_value REAL,
			approved_by INTEGER,
			approved_at DATETIME,
			document_url TEXT,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (asset_id) REFERENCES assets(id)
		)`,
		`CREATE TABLE IF NOT EXISTS asset_label_templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			paper_size TEXT DEFAULT 'A4',
			layout TEXT,
			include_logo INTEGER DEFAULT 1,
			include_school_name INTEGER DEFAULT 1,
			include_qr_code INTEGER DEFAULT 1,
			include_inventory_code INTEGER DEFAULT 1,
			include_asset_name INTEGER DEFAULT 1,
			include_funding_source INTEGER DEFAULT 1,
			qr_size INTEGER DEFAULT 50,
			font_size INTEGER DEFAULT 10,
			is_default INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS asset_stock_opname (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_name TEXT NOT NULL,
			start_date DATE NOT NULL,
			end_date DATE,
			status TEXT DEFAULT 'ongoing',
			created_by INTEGER NOT NULL,
			notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS asset_stock_opname_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			opname_id INTEGER NOT NULL,
			asset_id INTEGER NOT NULL,
			expected_location_id INTEGER,
			actual_location_id INTEGER,
			expected_condition TEXT,
			actual_condition TEXT,
			status TEXT,
			scanned_by INTEGER,
			scanned_at DATETIME,
			notes TEXT,
			FOREIGN KEY (opname_id) REFERENCES asset_stock_opname(id),
			FOREIGN KEY (asset_id) REFERENCES assets(id),
			FOREIGN KEY (expected_location_id) REFERENCES asset_locations(id),
			FOREIGN KEY (actual_location_id) REFERENCES asset_locations(id)
		)`,
	}
	for _, ddl := range schema {
		if _, err := db.Exec(ddl); err != nil {
			log.Printf("schema create warning: %v", err)
		}
	}

	// === Column additions for legacy DBs ===
	migrations := []struct {
		table  string
		column string
		alter  string
	}{
		{"attendance_logs", "method", "ALTER TABLE attendance_logs ADD COLUMN method TEXT DEFAULT 'RFID'"},
		{"attendance_logs", "note", "ALTER TABLE attendance_logs ADD COLUMN note TEXT DEFAULT ''"},
		{"classes", "wa_group_id", "ALTER TABLE classes ADD COLUMN wa_group_id TEXT DEFAULT ''"},
		{"prayer_logs", "status", "ALTER TABLE prayer_logs ADD COLUMN status TEXT DEFAULT 'Hadir'"},
		{"students", "parent_name", "ALTER TABLE students ADD COLUMN parent_name TEXT DEFAULT ''"},
		{"prayer_logs", "recorded_by", "ALTER TABLE prayer_logs ADD COLUMN recorded_by TEXT DEFAULT 'RFID'"},
		{"students", "birthday", "ALTER TABLE students ADD COLUMN birthday TEXT DEFAULT ''"},
		{"students", "status", "ALTER TABLE students ADD COLUMN status TEXT DEFAULT 'active'"},
		{"students", "nis_siswa", "ALTER TABLE students ADD COLUMN nis_siswa TEXT DEFAULT ''"},
		{"students", "password", "ALTER TABLE students ADD COLUMN password TEXT DEFAULT ''"},
		{"attendance_settings", "point_claim_enabled", "ALTER TABLE attendance_settings ADD COLUMN point_claim_enabled TEXT DEFAULT 'true'"},
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
	// === Core schema: create all tables IF NOT EXISTS (MySQL syntax) ===
	schema := []string{
		`CREATE TABLE IF NOT EXISTS schedules (
			id INT PRIMARY KEY AUTO_INCREMENT,
			time VARCHAR(10),
			label VARCHAR(255),
			audio_file VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS audio_files (
			id INT PRIMARY KEY AUTO_INCREMENT,
			file_name VARCHAR(255),
			display_name VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS majors (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS classes (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(255),
			major_id INT,
			wa_group_id VARCHAR(255) DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS students (
			id INT PRIMARY KEY AUTO_INCREMENT,
			rfid_uid VARCHAR(100) UNIQUE,
			nis VARCHAR(50),
			name VARCHAR(255),
			parent_phone VARCHAR(30),
			class_id INT,
			photo VARCHAR(255) DEFAULT '',
			parent_name VARCHAR(255) DEFAULT '',
			birthday VARCHAR(20) DEFAULT '',
			status VARCHAR(20) DEFAULT 'active',
			nis_siswa VARCHAR(50) DEFAULT '',
			password VARCHAR(255) DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS staff (
			id INT PRIMARY KEY AUTO_INCREMENT,
			rfid_uid VARCHAR(100) UNIQUE,
			nip VARCHAR(50) UNIQUE,
			name VARCHAR(255),
			phone VARCHAR(30),
			role VARCHAR(50)
		)`,
		`CREATE TABLE IF NOT EXISTS attendance_logs (
			id INT PRIMARY KEY AUTO_INCREMENT,
			rfid_uid VARCHAR(100),
			user_name VARCHAR(255),
			user_type VARCHAR(20),
			status VARCHAR(30),
			timestamp DATETIME,
			date DATE,
			method VARCHAR(50) DEFAULT 'RFID'
		)`,
		`CREATE TABLE IF NOT EXISTS attendance_settings (
			setting_key VARCHAR(100) PRIMARY KEY,
			setting_value TEXT,
			point_claim_enabled VARCHAR(10) DEFAULT 'true'
		)`,
		`CREATE TABLE IF NOT EXISTS prayer_logs (
			id INT PRIMARY KEY AUTO_INCREMENT,
			rfid_uid VARCHAR(100),
			name VARCHAR(255),
			class_name VARCHAR(255),
			prayer_type VARCHAR(20),
			timestamp DATETIME,
			date DATE,
			status VARCHAR(50) DEFAULT 'Hadir',
			recorded_by VARCHAR(50) DEFAULT 'RFID'
		)`,
		`CREATE TABLE IF NOT EXISTS whatsapp_logs (
			id INT PRIMARY KEY AUTO_INCREMENT,
			target VARCHAR(30),
			message TEXT,
			status VARCHAR(20),
			response TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS student_points (
			id INT PRIMARY KEY AUTO_INCREMENT,
			student_id INT,
			rule_id INT,
			reward_id INT,
			points_change INT,
			description TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			recorded_by VARCHAR(50)
		)`,
		`CREATE TABLE IF NOT EXISTS devices (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(255),
			ip_address VARCHAR(50),
			status VARCHAR(20),
			last_sync VARCHAR(50)
		)`,
		`CREATE TABLE IF NOT EXISTS announcements (
			id INT PRIMARY KEY AUTO_INCREMENT,
			title VARCHAR(255),
			message TEXT,
			audio_file VARCHAR(255),
			scheduled_at DATETIME,
			played_at DATETIME,
			status VARCHAR(20),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS holidays (
			id INT PRIMARY KEY AUTO_INCREMENT,
			date DATE NOT NULL,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50),
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS school_settings (
			id INT PRIMARY KEY AUTO_INCREMENT,
			setting_key VARCHAR(100) UNIQUE NOT NULL,
			setting_value TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS operators (
			id INT PRIMARY KEY AUTO_INCREMENT,
			username VARCHAR(100) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			phone VARCHAR(30),
			photo VARCHAR(255),
			is_active TINYINT DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS running_texts (
			id INT PRIMARY KEY AUTO_INCREMENT,
			content TEXT,
			is_active TINYINT DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS signage_media (
			id INT PRIMARY KEY AUTO_INCREMENT,
			filename VARCHAR(255),
			file_type VARCHAR(20),
			duration INT DEFAULT 10,
			is_active TINYINT DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS point_rules (
			id INT PRIMARY KEY AUTO_INCREMENT,
			category VARCHAR(100),
			name VARCHAR(255),
			points INT,
			description TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS point_claims (
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
		)`,
		`CREATE TABLE IF NOT EXISTS point_rewards (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(255),
			points_cost INT,
			stock INT,
			description TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS english_quests (
			id INT PRIMARY KEY AUTO_INCREMENT,
			date DATE NOT NULL UNIQUE,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			quest_type VARCHAR(50) NOT NULL DEFAULT 'written',
			topic VARCHAR(255) DEFAULT '',
			vocabulary_words TEXT DEFAULT '',
			quiz_question TEXT DEFAULT '',
			quiz_choices TEXT DEFAULT '',
			quiz_answer VARCHAR(255) DEFAULT '',
			xp_reward INT DEFAULT 10,
			created_by VARCHAR(100) DEFAULT 'admin',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS english_submissions (
			id INT PRIMARY KEY AUTO_INCREMENT,
			student_id INT NOT NULL,
			quest_id INT NOT NULL,
			quest_type VARCHAR(50) NOT NULL,
			content TEXT DEFAULT '',
			audio_file VARCHAR(255) DEFAULT '',
			vocab_words TEXT DEFAULT '',
			quiz_answer VARCHAR(255) DEFAULT '',
			xp_earned INT DEFAULT 0,
			status VARCHAR(20) DEFAULT 'pending',
			feedback TEXT DEFAULT '',
			submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			reviewed_at DATETIME,
			reviewed_by VARCHAR(100) DEFAULT '',
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(quest_id) REFERENCES english_quests(id)
		)`,
		`CREATE TABLE IF NOT EXISTS english_streaks (
			id INT PRIMARY KEY AUTO_INCREMENT,
			student_id INT NOT NULL UNIQUE,
			current_streak INT DEFAULT 0,
			longest_streak INT DEFAULT 0,
			total_xp INT DEFAULT 0,
			last_submit_date DATE DEFAULT NULL,
			FOREIGN KEY(student_id) REFERENCES students(id)
		)`,
		`CREATE TABLE IF NOT EXISTS english_badges (
			id INT PRIMARY KEY AUTO_INCREMENT,
			code VARCHAR(50) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			icon VARCHAR(10) DEFAULT '🏅',
			condition_type VARCHAR(50) NOT NULL,
			condition_value INT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS english_student_badges (
			id INT PRIMARY KEY AUTO_INCREMENT,
			student_id INT NOT NULL,
			badge_id INT NOT NULL,
			earned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY(student_id, badge_id),
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(badge_id) REFERENCES english_badges(id)
		)`,
		`CREATE TABLE IF NOT EXISTS ai_settings (
			id INT PRIMARY KEY AUTO_INCREMENT,
			setting_key VARCHAR(100) UNIQUE NOT NULL,
			setting_value TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		// === DUAL-TRACK POINT SYSTEM (MySQL) ===
		`CREATE TABLE IF NOT EXISTS achievement_rules (
			id INT PRIMARY KEY AUTO_INCREMENT,
			code VARCHAR(20) NOT NULL UNIQUE,
			category VARCHAR(100) NOT NULL,
			category_code VARCHAR(10) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			points INT NOT NULL DEFAULT 0,
			min_points INT DEFAULT 0,
			max_points INT DEFAULT 0,
			is_active TINYINT DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS violation_rules (
			id INT PRIMARY KEY AUTO_INCREMENT,
			code VARCHAR(20) NOT NULL UNIQUE,
			category VARCHAR(100) NOT NULL,
			category_code VARCHAR(10) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			points_1 INT NOT NULL DEFAULT 0,
			points_2 INT NOT NULL DEFAULT 0,
			points_3 INT NOT NULL DEFAULT 0,
			is_active TINYINT DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS student_achievement_points (
			id INT PRIMARY KEY AUTO_INCREMENT,
			student_id INT NOT NULL,
			rule_id INT,
			points INT NOT NULL DEFAULT 0,
			description TEXT,
			recorded_by VARCHAR(100) DEFAULT 'Admin',
			academic_year VARCHAR(20),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(rule_id) REFERENCES achievement_rules(id)
		)`,
		`CREATE TABLE IF NOT EXISTS student_violation_points (
			id INT PRIMARY KEY AUTO_INCREMENT,
			student_id INT NOT NULL,
			rule_id INT,
			occurrence INT DEFAULT 1,
			points INT NOT NULL DEFAULT 0,
			description TEXT,
			recorded_by VARCHAR(100) DEFAULT 'Admin',
			academic_year VARCHAR(20),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(rule_id) REFERENCES violation_rules(id)
		)`,
		// === SARPRAS TABLES (MySQL) ===
		`CREATE TABLE IF NOT EXISTS asset_categories (
			id INT PRIMARY KEY AUTO_INCREMENT,
			code VARCHAR(50) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_funding_sources (
			id INT PRIMARY KEY AUTO_INCREMENT,
			code VARCHAR(50) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_locations (
			id INT PRIMARY KEY AUTO_INCREMENT,
			code VARCHAR(50) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50),
			capacity INT,
			pic_staff_id INT,
			status VARCHAR(20) DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS assets (
			id INT PRIMARY KEY AUTO_INCREMENT,
			inventory_code VARCHAR(100) NOT NULL UNIQUE,
			qr_code VARCHAR(255) UNIQUE,
			category_id INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			specification TEXT,
			brand VARCHAR(100),
			serial_number VARCHAR(100),
			acquisition_year INT,
			acquisition_date DATE,
			purchase_price DECIMAL(15,2),
			funding_source_id INT,
			location_id INT,
			pic_staff_id INT,
			condition VARCHAR(20) DEFAULT 'good',
			quantity INT DEFAULT 1,
			unit VARCHAR(20),
			notes TEXT,
			photo_url VARCHAR(255),
			is_borrowable TINYINT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			created_by INT,
			FOREIGN KEY (category_id) REFERENCES asset_categories(id),
			FOREIGN KEY (location_id) REFERENCES asset_locations(id),
			FOREIGN KEY (funding_source_id) REFERENCES asset_funding_sources(id),
			INDEX idx_category (category_id),
			INDEX idx_location (location_id),
			INDEX idx_condition (condition),
			INDEX idx_borrowable (is_borrowable)
		)`,
		`CREATE TABLE IF NOT EXISTS asset_borrowings (
			id INT PRIMARY KEY AUTO_INCREMENT,
			asset_id INT NOT NULL,
			borrower_type VARCHAR(20) NOT NULL,
			borrower_id INT NOT NULL,
			purpose TEXT NOT NULL,
			borrow_date DATETIME NOT NULL,
			due_date DATETIME NOT NULL,
			return_date DATETIME,
			status VARCHAR(20) DEFAULT 'pending',
			approved_by INT,
			approved_at DATETIME,
			rejection_reason TEXT,
			return_condition VARCHAR(20),
			return_notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (asset_id) REFERENCES assets(id),
			INDEX idx_status (status),
			INDEX idx_borrower (borrower_type, borrower_id),
			INDEX idx_dates (borrow_date, due_date)
		)`,
		`CREATE TABLE IF NOT EXISTS asset_maintenance_tickets (
			id INT PRIMARY KEY AUTO_INCREMENT,
			ticket_number VARCHAR(50) NOT NULL UNIQUE,
			asset_id INT,
			location_id INT,
			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			priority VARCHAR(20) DEFAULT 'medium',
			status VARCHAR(20) DEFAULT 'open',
			reported_by INT NOT NULL,
			reported_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			assigned_to INT,
			photo_urls TEXT,
			resolution_notes TEXT,
			resolved_at DATETIME,
			closed_at DATETIME,
			cost DECIMAL(15,2),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (asset_id) REFERENCES assets(id),
			FOREIGN KEY (location_id) REFERENCES asset_locations(id),
			INDEX idx_status (status),
			INDEX idx_priority (priority)
		)`,
		`CREATE TABLE IF NOT EXISTS asset_movements (
			id INT PRIMARY KEY AUTO_INCREMENT,
			asset_id INT NOT NULL,
			from_location_id INT,
			to_location_id INT NOT NULL,
			moved_by INT NOT NULL,
			moved_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			reason TEXT,
			notes TEXT,
			FOREIGN KEY (asset_id) REFERENCES assets(id),
			FOREIGN KEY (from_location_id) REFERENCES asset_locations(id),
			FOREIGN KEY (to_location_id) REFERENCES asset_locations(id)
		)`,
		`CREATE TABLE IF NOT EXISTS asset_disposals (
			id INT PRIMARY KEY AUTO_INCREMENT,
			asset_id INT NOT NULL,
			disposal_date DATE NOT NULL,
			reason VARCHAR(100) NOT NULL,
			description TEXT,
			book_value DECIMAL(15,2),
			disposal_value DECIMAL(15,2),
			approved_by INT,
			approved_at DATETIME,
			document_url VARCHAR(255),
			created_by INT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (asset_id) REFERENCES assets(id)
		)`,
		`CREATE TABLE IF NOT EXISTS asset_label_templates (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL,
			paper_size VARCHAR(20) DEFAULT 'A4',
			layout TEXT,
			include_logo TINYINT DEFAULT 1,
			include_school_name TINYINT DEFAULT 1,
			include_qr_code TINYINT DEFAULT 1,
			include_inventory_code TINYINT DEFAULT 1,
			include_asset_name TINYINT DEFAULT 1,
			include_funding_source TINYINT DEFAULT 1,
			qr_size INT DEFAULT 50,
			font_size INT DEFAULT 10,
			is_default TINYINT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_stock_opname (
			id INT PRIMARY KEY AUTO_INCREMENT,
			session_name VARCHAR(255) NOT NULL,
			start_date DATE NOT NULL,
			end_date DATE,
			status VARCHAR(20) DEFAULT 'ongoing',
			created_by INT NOT NULL,
			notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS asset_stock_opname_items (
			id INT PRIMARY KEY AUTO_INCREMENT,
			opname_id INT NOT NULL,
			asset_id INT NOT NULL,
			expected_location_id INT,
			actual_location_id INT,
			expected_condition VARCHAR(20),
			actual_condition VARCHAR(20),
			status VARCHAR(20),
			scanned_by INT,
			scanned_at DATETIME,
			notes TEXT,
			FOREIGN KEY (opname_id) REFERENCES asset_stock_opname(id),
			FOREIGN KEY (asset_id) REFERENCES assets(id),
			FOREIGN KEY (expected_location_id) REFERENCES asset_locations(id),
			FOREIGN KEY (actual_location_id) REFERENCES asset_locations(id)
		)`,
	}
	for _, ddl := range schema {
		if _, err := db.Exec(ddl); err != nil {
			log.Printf("schema create warning: %v", err)
		}
	}

	// === Column additions for legacy DBs ===
	migrations := []struct {
		table  string
		column string
		alter  string
	}{
		{"attendance_logs", "method", "ALTER TABLE attendance_logs ADD COLUMN method VARCHAR(50) DEFAULT 'RFID'"},
		{"attendance_logs", "note", "ALTER TABLE attendance_logs ADD COLUMN note TEXT DEFAULT ''"},
		{"classes", "wa_group_id", "ALTER TABLE classes ADD COLUMN wa_group_id VARCHAR(255) DEFAULT ''"},
		{"prayer_logs", "status", "ALTER TABLE prayer_logs ADD COLUMN status VARCHAR(50) DEFAULT 'Hadir'"},
		{"students", "parent_name", "ALTER TABLE students ADD COLUMN parent_name VARCHAR(255) DEFAULT ''"},
		{"prayer_logs", "recorded_by", "ALTER TABLE prayer_logs ADD COLUMN recorded_by VARCHAR(50) DEFAULT 'RFID'"},
		{"students", "birthday", "ALTER TABLE students ADD COLUMN birthday VARCHAR(20) DEFAULT ''"},
		{"students", "status", "ALTER TABLE students ADD COLUMN status VARCHAR(20) DEFAULT 'active'"},
		{"students", "nis_siswa", "ALTER TABLE students ADD COLUMN nis_siswa VARCHAR(50) DEFAULT ''"},
		{"students", "password", "ALTER TABLE students ADD COLUMN password VARCHAR(255) DEFAULT ''"},
		{"attendance_settings", "point_claim_enabled", "ALTER TABLE attendance_settings ADD COLUMN point_claim_enabled VARCHAR(10) DEFAULT 'true'"},
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

func SeedEnglishBadges(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM english_badges").Scan(&count)
	if count > 0 {
		return
	}
	badges := []struct {
		code, name, desc, icon, condType string
		condVal                          int
	}{
		{"first_submit", "First Step", "Submit setoran pertama", "🌱", "total_submissions", 1},
		{"streak_3", "3-Day Streak", "3 hari berturut-turut", "🔥", "streak", 3},
		{"streak_7", "Week Warrior", "7 hari berturut-turut", "⚡", "streak", 7},
		{"streak_30", "Monthly Master", "30 hari berturut-turut", "👑", "streak", 30},
		{"xp_100", "XP Hunter", "Kumpulkan 100 XP", "💎", "total_xp", 100},
		{"xp_500", "XP Legend", "Kumpulkan 500 XP", "🏆", "total_xp", 500},
		{"vocab_10", "Word Collector", "Submit 10 vocabulary", "📚", "vocab_submissions", 10},
		{"quiz_10", "Quiz Champion", "Jawab 10 quiz dengan benar", "🎯", "quiz_correct", 10},
		{"writing_10", "Writer", "Submit 10 tulisan", "✍️", "writing_submissions", 10},
	}
	for _, b := range badges {
		db.Exec(`INSERT OR IGNORE INTO english_badges (code, name, description, icon, condition_type, condition_value) VALUES (?, ?, ?, ?, ?, ?)`,
			b.code, b.name, b.desc, b.icon, b.condType, b.condVal)
	}
}

func SeedAttendanceInsights(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM attendance_insights").Scan(&count)
	if count > 0 {
		return
	}
	
	insights := []struct {
		category string
		message  string
	}{
		// Early messages
		{"early", "Kamu lebih cepat {minutes} menit dari hari kemarin. Pertahankan! 🎉"},
		{"early", "Hebat! Kamu datang {minutes} menit lebih awal. Good job! 👏"},
		{"early", "Wow! {minutes} menit lebih cepat dari kemarin. Amazing! ⚡"},
		{"early", "Keren! Hari ini kamu {minutes} menit lebih pagi. Keep going! 💪"},
		{"early", "Luar biasa! Lebih cepat {minutes} menit. Disiplin sekali! 🌟"},
		
		// On time messages
		{"ontime", "Tepat waktu seperti biasa. Keep it up! 💪"},
		{"ontime", "Konsisten! Terus pertahankan kedisiplinanmu. 🌟"},
		{"ontime", "Perfect timing! Kamu selalu on time. Excellent! ✨"},
		{"ontime", "Disiplin adalah kunci kesuksesan. Great! 🎯"},
		{"ontime", "Selalu tepat waktu, kamu memang bisa diandalkan! 💯"},
		
		// Late messages
		{"late", "Kamu terlambat {minutes} menit hari ini. Besok lebih awal ya! ⏰"},
		{"late", "Ayo, besok datang lebih pagi! Kamu bisa! 💪"},
		{"late", "Terlambat {minutes} menit. Mari perbaiki besok! 🚀"},
		{"late", "Besok coba berangkat lebih pagi ya. Semangat! ⚡"},
		{"late", "Jangan sampai terlambat lagi. Kamu pasti bisa! 🎯"},
		
		// Sick messages
		{"sick", "Semoga lekas sembuh ya! 🏥💚"},
		{"sick", "Istirahat yang cukup dan jaga kesehatan. Get well soon! 🌈"},
		{"sick", "Semoga cepat pulih dan bisa kembali sekolah. Stay strong! 💪"},
		{"sick", "Jaga kesehatan dan minum obat teratur. Semoga cepat sehat! 🙏"},
		{"sick", "Sakit memang tidak enak. Semoga segera sembuh! 💖"},
		
		// Permission messages
		{"permission", "Ada keperluan hari ini. Semoga lancar! 🙏"},
		{"permission", "Semoga urusannya berjalan lancar. 📝"},
		{"permission", "Take care dan sampai jumpa besok! 👋"},
		{"permission", "Semoga keperluan hari ini berjalan baik. See you! 🌟"},
		{"permission", "Hati-hati di jalan. Sampai jumpa lagi! 🚗"},
	}
	
	for _, ins := range insights {
		db.Exec(`INSERT INTO attendance_insights (category, message, is_active) VALUES (?, ?, 1)`,
			ins.category, ins.message)
	}
}
