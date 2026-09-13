# Phase 3: Sistem Skor Komposit - Jalur Pemulihan Pelanggaran

## Overview
Implementasi Phase 3 dari sistem "3 REKA + 1 KONVERSI": sistem pelanggaran rehabilitatif dengan eskalasi otomatis, jalur pemulihan (redemption), dan SP1/SP2/SP3 sebagai early warning konseling.

## Konsep Pelanggaran Rehabilitatif

```
[Catat Pelanggaran] → [Eskalasi Otomatis 1→2→3] → [SP1/SP2/SP3 Warning]
         ↓
[Jalur Pemulihan] → [Kerja Positif] → [Verifikasi] → [Lunasi Poin]
         ↓
[Recalculate Burden] → [Update Composite Score]
```

## Eskalasi Pelanggaran

Setiap pelanggaran yang sama akan mengalami eskalasi:

| Pelanggaran ke- | Level | Poin | Contoh (P1-01: Terlambat <10min) |
|-----------------|-------|------|-----------------------------------|
| 1st             | 1     | points_1 | 5 poin |
| 2nd             | 2     | points_2 | 10 poin |
| 3rd+            | 3     | points_3 | 15 poin |

Sistem otomatis cek pelanggaran sebelumnya dengan rule_id yang sama di semester yang sama.

## SP1/SP2/SP3 — Early Warning Konseling

Bukan vonis, tapi ajakan untuk pembicaraan:

| Level | Threshold | Action |
|-------|-----------|--------|
| SP1   | 25 poin   | Surat peringatan + nasihat wali kelas |
| SP2   | 51 poin   | Surat peringatan kedua + konseling individual |
| SP3   | 76 poin   | Pemanggilan orang tua untuk konseling |

## Jalur Pemulihan (Redemption)

Pelanggaran bisa "dilunasi" lewat kerja positif dengan konversi **1:1**:

### Jenis Kerja Positif:
- **Piket ekstra** - membersihkan kelas/sekolah
- **Mentoring adik kelas** - bantu adik kelas pelajaran
- **Kerja sosial** - aksi sosial sekolah
- **Lainnya** - kegiatan positif lain yang disetujui

### Flow Pemulihan:
1. **Submit** - Guru input kerja positif siswa → `violation_redemptions` (status: pending)
2. **Verify** - Admin/pembina verify → approve/reject
3. **Apply** - Jika approve: update `redemption_points` di `student_violation_points` dan `semester_scores`
4. **Status** - `redemption_status` berubah:
   - `none` → `partially_redeemed` → `fully_redeemed`

## Rumus Beban Pelanggaran (Violation Burden)

```
Net Violation = Violation Points - Redemption Points

Violation Burden = (Net Violation / SP3_THRESHOLD) × 100

Max = 100
```

Contoh:
- Violation: 30 poin, Redemption: 10 poin → Net: 20
- Burden = (20 / 76) × 100 = 26.3

## Update Skor Komposit

Setiap perubahan pelanggaran/pemulihan auto-recalculate:

```
Composite = (0.30 × Attendance Index) 
          + (0.50 × Achievement Normalized) 
          + (0.20 × (100 - Violation Burden))
```

## API Endpoints

### 1. Catat Pelanggaran
**`POST /admin/scoring/violation/record`**

```json
{
  "student_id": 1,
  "rule_id": 5,
  "description": "Terlambat 15 menit",
  "recorded_by": "Pak Budi"
}
```

Response:
```json
{
  "status": "success",
  "message": "Pelanggaran dicatat: Terlambat 10-20 Menit (+15 poin)",
  "data": {
    "violation_id": 1,
    "rule_code": "P1-02",
    "rule_name": "Terlambat 10-20 Menit",
    "category": "Terlambat",
    "escalation_level": 2,
    "points": 15,
    "total_violation": 25,
    "sp_warning": {
      "level": "SP1",
      "threshold": 25,
      "recommendation": "Surat peringatan pertama + nasihat wali kelas"
    }
  }
}
```

### 2. Get Violation History
**`GET /admin/scoring/violation/list?student_id=1&semester_id=1`**

### 3. Submit Pemulihan
**`POST /admin/scoring/redemption/submit`**

```json
{
  "violation_id": 1,
  "student_id": 1,
  "redemption_type": "piket",
  "redemption_description": "Piket membersihkan kelas X-MPLB",
  "points_to_redeem": 10,
  "verified_by": "Pak Budi"
}
```

### 4. Verify Pemulihan
**`POST /admin/scoring/redemption/verify`**

```json
{
  "redemption_id": 1,
  "action": "approve",
  "verified_by": "Pak Budi"
}
```

### 5. Get Pending Pemulihan
**`GET /admin/scoring/redemption/pending?status=pending`**

## Tabel yang Digunakan

### `student_violation_points` (dari Phase 2)
- `escalation_level`: Level 1/2/3
- `redemption_status`: `none` / `partially_redeemed` / `fully_redeemed`
- `redemption_points`: Poin yang sudah dilunasi

### `violation_redemptions` (dari Phase 1)
- `redemption_type`: Jenis kerja positif
- `points_redeemed`: Poin yang dilunasi
- `status`: `pending` / `approved` / `rejected`
- `verified_by`, `verified_at`: Approval tracking

## Violation Rules (5 Kategori)

| Code | Kategori | Eskalasi Poin |
|------|----------|---------------|
| P1 | Terlambat | 5→10→15 |
| P2 | Kehadiran | 5→15→25 |
| P3 | Seragam | 5→15→25 |
| P4 | Kerapian | 5→13→15 |
| P5 | Kedisiplinan Umum | 5→28→30 |

## Files Changed
1. `internal/handler/violation_redemption.go` - Handler pelanggaran & pemulihan
2. `internal/router/router.go` - 5 routes baru

## Audit Trail
Setiap pelanggaran dicatat di `point_audit_log`:
- Action: `VIOLATION_RECORDED`
- Table: `student_violation_points`
- Performed_by: Yang mencatat
- Reason: Detail pelanggaran

## Next Steps (Phase 4)
- Dashboard leaderboard skor komposit
- UI untuk catat pelanggaran & pemulihan
- Integrasi WA untuk SP notification
- Reset semester otomatis
