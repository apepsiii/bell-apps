package repository

import (
	"database/sql"
	"log"
	"time"
)

// SeedAssetData seeds initial data for Sarpras module
func SeedAssetData(db *sql.DB) {
	SeedAssetCategories(db)
	SeedAssetFundingSources(db)
}

// SeedAssetCategories seeds default asset categories (KIB classification)
func SeedAssetCategories(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM asset_categories").Scan(&count)
	if count > 0 {
		return
	}

	categories := []struct {
		code, name, description string
	}{
		{"KIB-A", "Tanah", "Klasifikasi Inventaris Barang A: Tanah"},
		{"KIB-B", "Peralatan dan Mesin", "Klasifikasi Inventaris Barang B: Peralatan dan Mesin"},
		{"KIB-C", "Gedung dan Bangunan", "Klasifikasi Inventaris Barang C: Gedung dan Bangunan"},
		{"KIB-D", "Jalan, Irigasi, dan Jaringan", "Klasifikasi Inventaris Barang D: Jalan, Irigasi, dan Jaringan"},
		{"KIB-E", "Aset Tetap Lainnya", "Klasifikasi Inventaris Barang E: Aset Tetap Lainnya (koleksi perpustakaan, barang bercorak kesenian, hewan, ikan, tanaman)"},
		{"BHP", "Barang Habis Pakai", "Barang yang sekali pakai habis (ATK, bahan praktikum, dll)"},
	}

	for _, cat := range categories {
		_, err := db.Exec(`
			INSERT INTO asset_categories (code, name, description, created_at)
			VALUES (?, ?, ?, ?)
		`, cat.code, cat.name, cat.description, time.Now())
		if err != nil {
			log.Printf("Failed to seed asset category %s: %v", cat.code, err)
		}
	}

	log.Println("Asset categories seeded successfully")
}

// SeedAssetFundingSources seeds default funding sources
func SeedAssetFundingSources(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM asset_funding_sources").Scan(&count)
	if count > 0 {
		return
	}

	sources := []struct {
		code, name, description string
	}{
		{"BOS", "Dana BOS", "Dana Bantuan Operasional Sekolah dari pemerintah"},
		{"BOSDA", "Dana BOSDA", "Dana Bantuan Operasional Sekolah Daerah"},
		{"KOMITE", "Dana Komite", "Dana dari Komite Sekolah atau kontribusi orang tua siswa"},
		{"HIBAH", "Hibah", "Bantuan atau sumbangan dari pihak ketiga (CSR, alumni, donatur)"},
		{"APBD", "APBD", "Anggaran Pendapatan dan Belanja Daerah"},
		{"APBN", "APBN", "Anggaran Pendapatan dan Belanja Negara"},
		{"SWADAYA", "Swadaya", "Dari usaha sekolah sendiri (kantin, koperasi, dll)"},
	}

	for _, src := range sources {
		_, err := db.Exec(`
			INSERT INTO asset_funding_sources (code, name, description, created_at)
			VALUES (?, ?, ?, ?)
		`, src.code, src.name, src.description, time.Now())
		if err != nil {
			log.Printf("Failed to seed funding source %s: %v", src.code, err)
		}
	}

	log.Println("Asset funding sources seeded successfully")
}
