# Phase 1: Sistem Skor Komposit - Indeks Kehadiran

## Overview
Implementasi Phase 1 dari sistem "3 REKA + 1 KONVERSI": tabel dasar untuk skor komposit per semester dan fungsi hitung Indeks Kehadiran (0-100).

## Tabel Baru

### 1. `semesters`
Menyimpan data semester (Ganjil/Genap) dengan tanggal mulai dan selesai.

### 2. `semester_scores`
Menyimpan skor komposit per siswa per semester:
- `attendance_index` (DECIMAL 0-100): Indeks Kehadiran
- `attendance_days`, `late_days`, `absent_days`: Statistik kehadiran
- `streak_days`, `streak_bonus`: Bonus streak
- `achievement_points`, `achievement_normalized`: Poin prestasi (Phase 2)
- `violation_points`, `violation_burden`, `redemption_points`: Pelanggaran (Phase 3)
- `composite_score`: Skor komposit final
- UNIQUE(student_id, semester_id): satu skor per siswa per semester

### 3. `attendance_streaks`
Tracking streak (hari hadir berturut-turut):
- `current_streak`: Streak saat ini
- `longest_streak`: Streak terpanjang
- `last_attendance_date`: Tanggal hadir terakhir

### 4. `violation_redemptions`
Jalur pemulihan pelanggaran (Phase 3, tabel disiapkan):
- `redemption_type`: Jenis kerja positif
- `points_redeemed`: Poin yang dilunasi
- `verified_by`, `status`: Approval workflow

## Config Baru (point_config)
- `ATTENDANCE_WEIGHT` = 30
- `ACHIEVEMENT_WEIGHT` = 50
- `VIOLATION_WEIGHT` = 20
- `STREAK_BONUS_DAYS` = 10
- `STREAK_BONUS_POINTS` = 10
- `MAX_POINTS_PER_CATEGORY` = 200

## Rumus Indeks Kehadiran

```
Indeks = (Hadir × 100 + Terlambat × 50) / Total Hari Kerja

Jika CurrentStreak >= 10 hari:
    Indeks += 10 (bonus streak)

Max Indeks = 100
```

## API Endpoint

### `GET /admin/scoring/attendance-index`

**Parameters (query string):**
- `student_id` (optional): Hitung untuk siswa tertentu. Jika kosong, hitung untuk semua siswa.
- `semester_id` (optional): ID semester. Default: 1 (semester aktif).

**Response:**
```json
{
  "status": "success",
  "message": "15 indeks kehadiran berhasil dihitung",
  "results": [
    {
      "student_id": 1,
      "student_name": "ABDULAH FATHUR RAHMAN",
      "semester_id": 1,
      "attendance_days": 40,
      "late_days": 3,
      "absent_days": 2,
      "total_working_days": 45,
      "current_streak": 12,
      "longest_streak": 15,
      "streak_bonus": 10.0,
      "attendance_index": 88.89,
      "last_calculated_at": "2026-09-13 15:30:00"
    }
  ]
}
```

## Files Changed
1. `internal/repository/migration_semester_scores.go` - Migrasi tabel baru
2. `internal/repository/db.go` - Panggil MigrationSemesterScores
3. `internal/handler/attendance_index.go` - Handler dan logika perhitungan
4. `internal/router/router.go` - Route baru

## Contoh Perhitungan

### Skenario 1: Siswa Rajin
- Total hari kerja: 40
- Hadir tepat waktu: 38
- Terlambat: 2
- Alpha: 0
- Current streak: 20

```
Indeks = (38 × 100 + 2 × 50) / 40 = 3900 / 40 = 97.5
Bonus streak (20 >= 10): +10
Indeks final = 97.5 + 10 = 107.5 → capped to 100
```

### Skenario 2: Siswa Sering Sakit
- Total hari kerja: 40
- Hadir tepat waktu: 25
- Terlambat: 5
- Sakit: 10
- Current streak: 3

```
Indeks = (25 × 100 + 5 × 50) / 40 = 2750 / 40 = 68.75
Bonus streak (3 < 10): +0
Indeks final = 68.75
```

## Next Steps (Phase 2)
- Workflow approval prestasi (`pending_achievement_points`)
- Hitung poin prestasi per kategori dengan paritas
- Normalisasi poin prestasi ke 0-100
