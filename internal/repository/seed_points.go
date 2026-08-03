package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// SeedDualTrackPointRules seeds achievement rules (R1-R10) and violation rules (P1-P5)
func SeedDualTrackPointRules(db *sql.DB) {
	seedAchievementRules(db)
	seedViolationRules(db)
}

func seedAchievementRules(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM achievement_rules").Scan(&count)
	if count > 0 {
		return
	}

	rules := []struct {
		code, category, categoryCode, name, description string
		points, minPoints, maxPoints                    int
	}{
		// R1 - Pengembangan Keagamaan
		{"R1-01", "Pengembangan Keagamaan", "R1", "Praktik Keagamaan", "Melaksanakan praktik ibadah keagamaan", 20, 20, 20},

		// R2 - Kejujuran
		{"R2-01", "Kejujuran", "R2", "Melaporkan Barang Temuan", "Melaporkan barang temuan kepada yang berwenang", 10, 10, 10},
		{"R2-02", "Kejujuran", "R2", "Berkata Jujur", "Berkata jujur dalam situasi yang menentukan", 15, 10, 20},
		{"R2-03", "Kejujuran", "R2", "Jujur dalam Ujian", "Tidak menyontek dan berlaku jujur dalam ujian", 20, 10, 20},

		// R3 - Prestasi Akademik
		{"R3-01", "Prestasi Akademik", "R3", "Peringkat Kelas", "Meraih peringkat 1-3 di kelas", 10, 10, 10},
		{"R3-02", "Prestasi Akademik", "R3", "Siswa Aktif", "Aktif dalam kegiatan belajar mengajar", 10, 10, 10},
		{"R3-03", "Prestasi Akademik", "R3", "Prestasi Tingkat Sekolah", "Meraih prestasi akademik tingkat sekolah", 15, 15, 15},
		{"R3-04", "Prestasi Akademik", "R3", "Prestasi Tingkat Kota/Kabupaten", "Meraih prestasi akademik tingkat kota/kabupaten", 25, 25, 25},
		{"R3-05", "Prestasi Akademik", "R3", "Prestasi Tingkat Provinsi", "Meraih prestasi akademik tingkat provinsi", 35, 35, 35},
		{"R3-06", "Prestasi Akademik", "R3", "Prestasi Tingkat Nasional/Beasiswa", "Meraih prestasi akademik tingkat nasional atau mendapat beasiswa", 40, 40, 40},

		// R4 - Kedisiplinan
		{"R4-01", "Kedisiplinan", "R4", "Disiplin 3 Bulan Berturut-turut", "Tidak melanggar aturan selama 3 bulan berturut-turut", 10, 10, 10},
		{"R4-02", "Kedisiplinan", "R4", "Disiplin 6 Bulan Berturut-turut", "Tidak melanggar aturan selama 6 bulan berturut-turut", 20, 20, 20},
		{"R4-03", "Kedisiplinan", "R4", "Disiplin 9 Bulan Berturut-turut", "Tidak melanggar aturan selama 9 bulan berturut-turut", 35, 35, 35},
		{"R4-04", "Kedisiplinan", "R4", "Disiplin 12 Bulan Berturut-turut", "Tidak melanggar aturan selama 12 bulan berturut-turut", 50, 50, 50},

		// R5 - Pengembangan Sosial
		{"R5-01", "Pengembangan Sosial", "R5", "Menolong Korban Musibah", "Aktif menolong korban musibah atau bencana", 10, 10, 10},
		{"R5-02", "Pengembangan Sosial", "R5", "Aksi Sosial", "Terlibat aktif dalam aksi sosial sekolah", 15, 10, 15},

		// R6 - Kepemimpinan
		{"R6-01", "Kepemimpinan", "R6", "Mengikuti LDKS", "Mengikuti Latihan Dasar Kepemimpinan Siswa", 10, 10, 10},
		{"R6-02", "Kepemimpinan", "R6", "Pengurus Organisasi", "Menjadi pengurus OSIS, MPK, atau ekstrakurikuler", 15, 15, 15},
		{"R6-03", "Kepemimpinan", "R6", "Ketua Organisasi", "Menjadi ketua OSIS, MPK, atau ekstrakurikuler", 20, 20, 20},

		// R7 - Kebangsaan
		{"R7-01", "Kebangsaan", "R7", "Pelaksanaan Nilai Pancasila", "Aktif melaksanakan nilai-nilai Pancasila", 10, 10, 10},
		{"R7-02", "Kebangsaan", "R7", "Bela Negara", "Mengikuti kegiatan bela negara", 15, 15, 15},
		{"R7-03", "Kebangsaan", "R7", "Petugas Upacara", "Menjadi petugas upacara bendera", 10, 10, 10},
		{"R7-04", "Kebangsaan", "R7", "Duta Budaya Tingkat Kota", "Menjadi duta budaya tingkat kota/kabupaten", 20, 20, 20},
		{"R7-05", "Kebangsaan", "R7", "Duta Budaya Tingkat Provinsi", "Menjadi duta budaya tingkat provinsi", 30, 30, 30},
		{"R7-06", "Kebangsaan", "R7", "Duta Budaya Tingkat Nasional", "Menjadi duta budaya tingkat nasional", 40, 40, 40},

		// R8 - Ekstrakurikuler & Perlombaan
		{"R8-01", "Ekstrakurikuler & Perlombaan", "R8", "Aktif Ekstrakurikuler", "Aktif mengikuti kegiatan ekstrakurikuler", 5, 5, 5},
		{"R8-02", "Ekstrakurikuler & Perlombaan", "R8", "Juara Tingkat Sekolah", "Meraih juara dalam perlombaan tingkat sekolah", 10, 10, 10},
		{"R8-03", "Ekstrakurikuler & Perlombaan", "R8", "Juara Tingkat Kota/Kabupaten", "Meraih juara dalam perlombaan tingkat kota/kabupaten", 20, 20, 20},
		{"R8-04", "Ekstrakurikuler & Perlombaan", "R8", "Juara Tingkat Provinsi", "Meraih juara dalam perlombaan tingkat provinsi", 35, 35, 35},
		{"R8-05", "Ekstrakurikuler & Perlombaan", "R8", "Juara Tingkat Nasional", "Meraih juara dalam perlombaan tingkat nasional", 50, 50, 50},

		// R9 - Peduli Lingkungan
		{"R9-01", "Peduli Lingkungan", "R9", "Membuang/Memilah Sampah", "Aktif membuang dan memilah sampah pada tempatnya", 10, 10, 10},
		{"R9-02", "Peduli Lingkungan", "R9", "Karya Ramah Lingkungan", "Menciptakan karya atau inovasi ramah lingkungan", 15, 15, 15},
		{"R9-03", "Peduli Lingkungan", "R9", "Reboisasi", "Aktif dalam kegiatan reboisasi atau penghijauan", 20, 20, 20},

		// R10 - Kewirausahaan
		{"R10-01", "Kewirausahaan", "R10", "Ide Ekonomis", "Memberikan ide ekonomis yang bermanfaat", 10, 10, 10},
		{"R10-02", "Kewirausahaan", "R10", "Jiwa Wirausaha", "Mengembangkan jiwa kewirausahaan secara aktif", 10, 10, 10},
		{"R10-03", "Kewirausahaan", "R10", "Tabungan/Usaha Mandiri", "Memiliki tabungan atau usaha mandiri", 10, 10, 10},
	}

	for _, r := range rules {
		_, err := db.Exec(`
			INSERT INTO achievement_rules (code, category, category_code, name, description, points, min_points, max_points, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
		`, r.code, r.category, r.categoryCode, r.name, r.description, r.points, r.minPoints, r.maxPoints)
		if err != nil {
			log.Printf("Failed to seed achievement rule %s: %v", r.code, err)
		}
	}

	log.Println("Achievement rules seeded successfully")
}

func seedViolationRules(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM violation_rules").Scan(&count)
	if count > 0 {
		return
	}

	rules := []struct {
		code, category, categoryCode, name, description string
		p1, p2, p3                                      int
	}{
		// P1 - Terlambat
		{"P1-01", "Terlambat", "P1", "Terlambat < 10 Menit", "Keterlambatan masuk kurang dari 10 menit", 5, 10, 15},
		{"P1-02", "Terlambat", "P1", "Terlambat 10-20 Menit", "Keterlambatan masuk 10 hingga 20 menit", 10, 15, 20},
		{"P1-03", "Terlambat", "P1", "Terlambat 20-30 Menit", "Keterlambatan masuk 20 hingga 30 menit", 15, 17, 20},
		{"P1-04", "Terlambat", "P1", "Terlambat > 30 Menit", "Keterlambatan masuk lebih dari 30 menit", 20, 20, 20},

		// P2 - Kehadiran
		{"P2-01", "Kehadiran", "P2", "Izin Tanpa Keterangan", "Izin tanpa keterangan yang jelas", 5, 10, 15},
		{"P2-02", "Kehadiran", "P2", "Alpa/Tidak Hadir", "Tidak hadir tanpa keterangan (alpa)", 10, 15, 20},
		{"P2-03", "Kehadiran", "P2", "Memalsukan Surat", "Memalsukan surat izin atau keterangan", 20, 23, 25},
		{"P2-04", "Kehadiran", "P2", "Keluar Tanpa Izin", "Keluar lingkungan sekolah tanpa izin saat jam pelajaran", 10, 15, 20},

		// P3 - Seragam
		{"P3-01", "Seragam", "P3", "Atribut Tidak Lengkap", "Tidak menggunakan atribut seragam secara lengkap", 5, 10, 15},
		{"P3-02", "Seragam", "P3", "Baju Tidak Dimasukkan", "Baju tidak dimasukkan ke dalam celana/rok", 5, 10, 15},
		{"P3-03", "Seragam", "P3", "Memakai Jaket di Sekolah", "Memakai jaket di lingkungan sekolah tanpa izin", 10, 15, 20},
		{"P3-04", "Seragam", "P3", "Mencoret Seragam", "Mencoret atau merusak seragam sekolah", 20, 23, 25},

		// P4 - Kerapian dan Penampilan
		{"P4-01", "Kerapian dan Penampilan", "P4", "Kuku Panjang/Dicat", "Kuku panjang atau dicat", 5, 10, 15},
		{"P4-02", "Kerapian dan Penampilan", "P4", "Rambut Gondrong/Diwarnai", "Rambut gondrong, diwarnai, atau model tidak sesuai", 10, 13, 15},
		{"P4-03", "Kerapian dan Penampilan", "P4", "Bertato", "Memiliki tato di tubuh", 10, 13, 15},
		{"P4-04", "Kerapian dan Penampilan", "P4", "Perhiasan/Riasan Berlebih", "Memakai perhiasan atau riasan yang berlebihan", 5, 10, 15},
		{"P4-05", "Kerapian dan Penampilan", "P4", "Lensa Kontak Kosmetik", "Memakai lensa kontak kosmetik berwarna", 5, 10, 15},

		// P5 - Kedisiplinan Umum
		{"P5-01", "Kedisiplinan Umum", "P5", "Gawai Tanpa Izin", "Menggunakan HP tanpa izin saat KBM", 5, 10, 15},
		{"P5-02", "Kedisiplinan Umum", "P5", "Makan/Minum di Kelas", "Makan atau minum di dalam kelas saat KBM", 5, 10, 15},
		{"P5-03", "Kedisiplinan Umum", "P5", "Menyontek", "Menyontek saat ujian atau tugas", 15, 20, 25},
		{"P5-04", "Kedisiplinan Umum", "P5", "Menyebarkan Hoaks", "Menyebarkan berita bohong atau hoaks", 15, 20, 25},
		{"P5-05", "Kedisiplinan Umum", "P5", "Merokok", "Merokok di lingkungan sekolah", 20, 25, 30},
		{"P5-06", "Kedisiplinan Umum", "P5", "Narkoba", "Terbukti menggunakan atau mengedarkan narkoba", 30, 30, 30},
		{"P5-07", "Kedisiplinan Umum", "P5", "Pergaulan Bebas", "Terlibat dalam pergaulan bebas", 25, 28, 30},
		{"P5-08", "Kedisiplinan Umum", "P5", "Tawuran", "Terlibat dalam tawuran atau perkelahian", 25, 28, 30},
		{"P5-09", "Kedisiplinan Umum", "P5", "Vandalisme", "Melakukan vandalisme atau merusak fasilitas sekolah", 15, 20, 25},
		{"P5-10", "Kedisiplinan Umum", "P5", "Perundungan (Bullying)", "Melakukan perundungan terhadap sesama siswa", 20, 25, 30},
		{"P5-11", "Kedisiplinan Umum", "P5", "Membawa Senjata/Alat Judi", "Membawa senjata tajam atau alat perjudian", 25, 28, 30},
	}

	for _, r := range rules {
		_, err := db.Exec(`
			INSERT INTO violation_rules (code, category, category_code, name, description, points_1, points_2, points_3, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
		`, r.code, r.category, r.categoryCode, r.name, r.description, r.p1, r.p2, r.p3)
		if err != nil {
			log.Printf("Failed to seed violation rule %s: %v", r.code, err)
		}
	}

	log.Println("Violation rules seeded successfully")
}

// CurrentAcademicYear returns e.g. "2025/2026"
// Academic year starts in July
func CurrentAcademicYear() string {
	now := time.Now()
	year := now.Year()
	if now.Month() >= 7 {
		return fmt.Sprintf("%d/%d", year, year+1)
	}
	return fmt.Sprintf("%d/%d", year-1, year)
}
