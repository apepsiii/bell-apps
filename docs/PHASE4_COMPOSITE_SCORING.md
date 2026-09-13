# Phase 4: Dashboard Leaderboard & UI

## Overview
Implementasi Phase 4: UI lengkap untuk sistem skor komposit di admin panel dengan 4 tab baru.

## Tab Baru di Section "Sistem Poin Siswa"

### 1. Tab "Skor Komposit"
Leaderboard dengan skor komposit (bukan poin mentah):

- **Info Banner**: Menjelaskan rumus skor komposit + tombol "Hitung Ulang"
- **Filter Kelas**: Filter leaderboard per kelas
- **Tabel Leaderboard**: Rank, Nama, Kelas, Kehadiran, Prestasi, Pelanggaran (dengan redemption), Skor Komposit, Detail
- **Rank Badge**: Emas (1), Perak (2), Perunggu (3)
- **Detail Modal**: Popup dengan breakdown lengkap:
  - Kehadiran: indeks + statistik + streak bonus
  - Prestasi: normalized + total poin
  - Pelanggaran: burden + total + redeemed
  - Skor Komposit: besar + breakdown per komponen

### 2. Tab "Approval Prestasi"
Workflow approval prestasi dengan 2 kolom:

**Kiri - Ajukan Prestasi:**
- Search siswa (autocomplete)
- Pilih rule prestasi (grouped by kategori)
- Keterangan (textarea)
- URL bukti (auto-show untuk poin >= 50)
- Submit ke pending

**Kanan - Pending Approval:**
- List pengajuan pending dengan badge counter
- Setiap item: nama, rule, kategori, poin, bukti link
- Tombol: Setujui / Tolak (dengan alasan)

### 3. Tab "Pelanggaran"
Workflow pelanggaran dengan 2 kolom:

**Kiri - Catat Pelanggaran:**
- Search siswa (autocomplete)
- Pilih rule pelanggaran (grouped by kategori)
- Keterangan (textarea)
- Submit dengan eskalasi otomatis
- SP Warning popup jika threshold tercapai

**Kanan - Riwayat Pelanggaran:**
- List semua pelanggaran
- Badge eskalasi (1=yellow, 2=orange, 3=red)
- Status redemption (Lunas/Sebagian/-)

### 4. Tab "Pemulihan"
Approval pemulihan (kerja positif):

- List pengajuan pemulihan pending
- Setiap item: nama siswa, jenis kerja (piket/mentoring/sosial), poin
- Tombol: Setujui / Tolak
- Badge counter

## JavaScript Functions

### Leaderboard
- `loadKompositLeaderboard()` - Load data dari `/admin/scoring/composite`
- `recalcAllScores()` - Trigger recalculation via `/admin/scoring/attendance-index`
- `viewCompositeDetail(studentID)` - Popup detail breakdown

### Achievement
- `loadAchievementRules()` - Load rules dari `/admin/v2/achievement-rules`
- `searchStudentForPrestasi(query)` - Autocomplete search
- `submitAchievement(event)` - POST ke `/admin/scoring/achievement/submit`
- `loadPendingAchievements()` - GET dari `/admin/scoring/achievement/pending`
- `approveAchievement(id)` / `rejectAchievement(id)` - POST ke `/admin/scoring/achievement/approve`

### Violation
- `loadViolationRules()` - Load rules dari `/admin/v2/violation-rules`
- `searchStudentForViolation(query)` - Autocomplete search
- `submitViolation(event)` - POST ke `/admin/scoring/violation/record`
- `loadViolationHistory()` - GET dari `/admin/scoring/violation/list`

### Redemption
- `loadPendingRedemptions()` - GET dari `/admin/scoring/redemption/pending`
- `verifyRedemption(id, action)` - POST ke `/admin/scoring/redemption/verify`

## API Endpoints Used

| Endpoint | Method | Fungsi |
|----------|--------|--------|
| `/admin/scoring/composite` | GET | Leaderboard skor komposit |
| `/admin/scoring/attendance-index` | GET | Hitung ulang indeks kehadiran |
| `/admin/scoring/achievement/submit` | POST | Ajukan prestasi |
| `/admin/scoring/achievement/approve` | POST | Approve/reject prestasi |
| `/admin/scoring/achievement/pending` | GET | List pending prestasi |
| `/admin/scoring/violation/record` | POST | Catat pelanggaran |
| `/admin/scoring/violation/list` | GET | Riwayat pelanggaran |
| `/admin/scoring/redemption/submit` | POST | Ajukan pemulihan |
| `/admin/scoring/redemption/verify` | POST | Verify pemulihan |
| `/admin/scoring/redemption/pending` | GET | List pending pemulihan |
| `/admin/v2/achievement-rules` | GET | List achievement rules |
| `/admin/v2/violation-rules` | GET | List violation rules |
| `/admin/points/search-student` | GET | Search siswa |

## Files Changed
1. `views/admin.html` - 4 tab baru + JavaScript functions

## UI Flow

```
[Skor Komposit] → Lihat leaderboard → Klik "Detail" → Popup breakdown
                                                      ↓
[Approval Prestasi] → Submit prestasi → Pending list → Approve/Reject
                                                           ↓
[Pelanggaran] → Catat pelanggaran → Eskalasi otomatis → SP Warning
                                     ↓
[Pemulihan] → Submit kerja positif → Pending list → Verify → Lunasi poin
```

## Testing Checklist
- [ ] Tab "Skor Komposit" load data
- [ ] Filter per kelas berfungsi
- [ ] Tombol "Hitung Ulang" berfungsi
- [ ] Popup detail menampilkan breakdown
- [ ] Submit prestasi tersimpan ke pending
- [ ] Approve prestasi pindah ke student_points
- [ ] Reject prestasi dengan alasan
- [ ] Auto-show proof field untuk poin >= 50
- [ ] Catat pelanggaran dengan eskalasi
- [ ] SP1/SP2/SP3 warning popup
- [ ] Submit pemulihan tersimpan
- [ ] Verify pemulihan update redemption
- [ ] Riwayat pelanggaran menampilkan status redemption

## Complete System Summary (Phase 1-4)

### Rumus Final
```
Skor Komposit = (0.30 × Indeks Kehadiran) 
             + (0.50 × Poin Prestasi ternormalisasi) 
             + (0.20 × (100 - Beban Pelanggaran))
```

### Komponen
1. **Indeks Kehadiran (0-100)**: Hadir=100, Telat=50, Alpha=0, Streak bonus=+10
2. **Poin Prestasi (0-100)**: 10 kategori, max 200/kategori, approval workflow
3. **Beban Pelanggaran (0-100)**: Eskalasi 1/2/3, redemption 1:1, SP1/SP2/SP3

### Config
| Key | Value |
|-----|-------|
| ATTENDANCE_WEIGHT | 30 |
| ACHIEVEMENT_WEIGHT | 50 |
| VIOLATION_WEIGHT | 20 |
| STREAK_BONUS_DAYS | 10 |
| STREAK_BONUS_POINTS | 10 |
| MAX_POINTS_PER_CATEGORY | 200 |
| APPROVAL_THRESHOLD | 50 |
| SP1_THRESHOLD | 25 |
| SP2_THRESHOLD | 51 |
| SP3_THRESHOLD | 76 |
