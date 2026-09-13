# Phase 2: Sistem Skor Komposit - Workflow Approval Prestasi

## Overview
Implementasi Phase 2 dari sistem "3 REKA + 1 KONVERSI": workflow approval prestasi dengan paritas kategori dan normalisasi poin prestasi ke 0-100.

## Tabel Baru

### `student_achievement_points`
Menyimpan prestasi yang sudah disetujui:
- `student_id`, `rule_id`, `pending_id`: Relasi
- `points`: Poin yang diberikan
- `approved_by`, `approved_at`: Approval tracking
- `semester_id`: Semester terkait
- Soft delete: `deleted_at`, `deleted_by`, `deletion_reason`

### `student_violation_points`
Disiapkan untuk Phase 3:
- `escalation_level`: Level pelanggaran (1/2/3)
- `redemption_status`: Status pemulihan
- `redemption_points`: Poin yang sudah dilunasi

### Kolom Baru di `pending_achievement_points`
- `proof_url`: URL foto/sertifikat bukti
- `semester_id`: Semester terkait

## Workflow Approval

```
[Submit Prestasi] → [Pending Review] → [Approve/Reject] → [student_points + semester_scores]
```

1. **Submit**: Siswa/Guru input prestasi → masuk `pending_achievement_points` dengan status `pending`
2. **Review**: Guru lihat daftar pending di `/scoring/achievement/pending`
3. **Approve**: Poin dipindahkan ke `student_points` + update `semester_scores.achievement_points`
4. **Reject**: Status berubah `rejected` dengan alasan

## Validasi

### APPROVAL_THRESHOLD (= 50)
Poin >= 50 **wajib** melampirkan bukti (foto/sertifikat):
- `proof_url` tidak boleh kosong

### Paritas Kategori (MAX_POINTS_PER_CATEGORY = 200)
Setiap kategori (R1-R10) maksimal 200 poin per semester:
- Cek total poin per kategori sebelum approve
- Jika melebihi, reject dengan pesan error

## Normalisasi Poin Prestasi

```
Normalized = (Total Achievement Points / Max Possible) × 100

Max Possible = 10 kategori × 200 poin = 2000

Jika > 100, cap ke 100
```

## API Endpoints

### 1. Submit Prestasi
**`POST /admin/scoring/achievement/submit`**

Request:
```json
{
  "student_id": 1,
  "rule_id": 5,
  "description": "Juara 1 lomba cerdas cermat tingkat provinsi",
  "proof_url": "/uploads/certificates/2026/01.jpg",
  "requested_by": "Pak Budi"
}
```

Response:
```json
{
  "status": "success",
  "message": "Pengajuan prestasi berhasil dikirim untuk approval",
  "data": {
    "rule_code": "R8-04",
    "rule_name": "Juara Tingkat Provinsi",
    "category": "Ekstrakurikuler & Perlombaan",
    "points": 35,
    "description": "Juara 1 lomba cerdas cermat tingkat provinsi"
  }
}
```

### 2. Approve/Reject Prestasi
**`POST /admin/scoring/achievement/approve`**

Request:
```json
{
  "pending_id": 1,
  "action": "approve",
  "approved_by": "Pak Budi"
}
```

Atau reject:
```json
{
  "pending_id": 1,
  "action": "reject",
  "approved_by": "Pak Budi",
  "rejection_reason": "Bukti tidak valid"
}
```

### 3. Get Pending Achievements
**`GET /admin/scoring/achievement/pending?status=pending`**

Response:
```json
{
  "status": "success",
  "total": 3,
  "results": [...]
}
```

### 4. Get Composite Scores (Leaderboard)
**`GET /admin/scoring/composite`**

Parameters:
- `semester_id` (optional, default: 1)
- `class_id` (optional, filter per kelas)
- `student_id` (optional, untuk detail 1 siswa)

Response:
```json
{
  "status": "success",
  "total": 15,
  "results": [
    {
      "student_id": 1,
      "student_name": "ABDULAH FATHUR RAHMAN",
      "nis": "128777",
      "class_name": "X-MPLB",
      "rank": 1,
      "attendance_index": 95.5,
      "achievement_points": 120,
      "achievement_normalized": 6.0,
      "violation_points": 0,
      "violation_burden": 0,
      "composite_score": 31.65
    }
  ]
}
```

## Achievement Rules (10 Kategori)

| Code | Kategori | Max Poin |
|------|----------|----------|
| R1 | Pengembangan Keagamaan | 20 |
| R2 | Kejujuran | 20 |
| R3 | Prestasi Akademik | 40 |
| R4 | Kedisiplinan | 50 |
| R5 | Pengembangan Sosial | 15 |
| R6 | Kepemimpinan | 20 |
| R7 | Kebangsaan | 40 |
| R8 | Ekstrakurikuler & Perlombaan | 50 |
| R9 | Peduli Lingkungan | 20 |
| R10 | Kewirausahaan | 10 |

## Files Changed
1. `internal/repository/migration_semester_scores.go` - Tambah tabel dan kolom
2. `internal/handler/achievement_approval.go` - Submit & approve handler
3. `internal/handler/composite_score.go` - Leaderboard & normalisasi
4. `internal/router/router.go` - Routes baru

## Next Steps (Phase 3)
- Jalur pemulihan pelanggaran (redemption)
- Hitung violation_burden
- SP1/SP2/SP3 sebagai early warning konseling
- Update composite_score dengan violation component
