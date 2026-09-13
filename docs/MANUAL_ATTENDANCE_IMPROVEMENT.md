# Manual Attendance System Improvement

## Overview
Peningkatan sistem absensi manual dengan opsi status yang lebih lengkap dan pencatatan waktu yang fleksibel.

## Status Kehadiran Baru

### 1. **Hadir**
- Set jam kehadiran sebelum batas jam masuk
- Input: waktu kehadiran (time picker)
- Behavior: Catat sebagai "Hadir" dengan timestamp custom

### 2. **Terlambat**
- Set jam kehadiran setelah batas jam masuk
- Input: waktu kehadiran (time picker)
- Behavior: Catat sebagai "Terlambat" dengan timestamp custom

### 3. **Sakit**
- Dua opsi radio button:
  - **Dengan Surat Keterangan**: Radio "Dengan Surat Keterangan" dipilih
  - **Tanpa Surat Keterangan**: Radio "Tanpa Surat Keterangan" dipilih
- Input: keterangan (textarea, opsional) + radio button pilihan surat
- Status disimpan sebagai:
  - `Sakit (Dengan Surat)` jika radio "with" dipilih
  - `Sakit (Tanpa Surat)` jika radio "without" dipilih
  - `Sakit` jika tidak ada radio yang dipilih

### 4. **Dispensasi**
- Izin dengan keterangan wajib
- Input: textarea untuk keterangan (required)
- Behavior: Catat sebagai "Dispensasi" dengan note

### 5. **Alpha**
- Tanpa keterangan
- Tidak ada input tambahan
- Behavior: Catat sebagai "Alpha"

## Database Changes

### New Column: `attendance_logs.note`
```sql
-- SQLite
ALTER TABLE attendance_logs ADD COLUMN note TEXT DEFAULT '';

-- MySQL
ALTER TABLE attendance_logs ADD COLUMN note TEXT DEFAULT '';
```

Kolom ini menyimpan:
- Keterangan sakit
- Keterangan dispensasi
- Informasi tambahan lainnya

## API Changes

### Endpoint: `POST /admin/attendance/manual`

**Request Parameters:**
```
student_id: string (required)
status: string (required) - "Hadir", "Terlambat", "Sakit", "Dispensasi", "Alpha"
time: string (optional) - Format "HH:MM", required untuk Hadir/Terlambat
note: string (optional) - Keterangan, required untuk Dispensasi
letter_status: string (optional) - "with" atau "without", hanya untuk status Sakit
```

**Response:**
```json
{
  "status": "success",
  "message": "Absensi manual berhasil disimpan"
}
```

## UI Changes

### Modal Form
- **Grid 3 kolom pertama**: Hadir, Terlambat, Alpha
- **Grid 2 kolom kedua**: Sakit, Dispensasi
- **Conditional Sections**:
  - Time input: tampil untuk Hadir/Terlambat
  - Note textarea: tampil untuk Sakit/Dispensasi
  - Radio button (Dengan/Tanpa Surat): tampil untuk Sakit

### Form Validation
- Status wajib dipilih
- Waktu wajib diisi untuk Hadir/Terlambat
- Keterangan wajib diisi untuk Dispensasi
- Keterangan opsional untuk Sakit
- Radio button surat opsional untuk Sakit

## WhatsApp Notification

Notifikasi WA hanya dikirim untuk status:
- **Hadir**: Gunakan template `wa_template_in`
- **Terlambat**: Gunakan template `wa_template_late`

Status lain (Sakit, Dispensasi, Alpha) tidak mengirim notifikasi WA otomatis.

## Migration
Migrasi kolom `note` dilakukan otomatis saat aplikasi start melalui:
- `runSQLiteMigrations()` untuk SQLite
- `runMySQLMigrations()` untuk MySQL

## Files Changed
1. `views/admin.html` - UI modal, JavaScript, dan status colors
2. `main.go` - Handler `ManualAttendanceHandler`
3. `internal/repository/db.go` - Migration untuk kolom `note`
4. `views/profile.html` - Kalender legend dan render status baru
5. `docs/MANUAL_ATTENDANCE_IMPROVEMENT.md` - Dokumentasi

## Testing Checklist
- [ ] Absen Hadir dengan jam custom
- [ ] Absen Terlambat dengan jam custom
- [ ] Sakit tanpa surat keterangan
- [ ] Sakit dengan surat keterangan
- [ ] Sakit tanpa pilih radio (hanya "Sakit")
- [ ] Dispensasi dengan keterangan (wajib)
- [ ] Alpha tanpa keterangan
- [ ] Validasi form (wajib waktu, wajib keterangan)
- [ ] WA notification untuk Hadir/Terlambat
- [ ] Data tersimpan dengan benar di DB
- [ ] Tampilan status baru di kalender profile
- [ ] Tampilan status baru di laporan admin

## Future Improvements
- [ ] History log untuk melihat keterangan
- [ ] Upload attachment untuk surat keterangan
- [ ] Approval workflow untuk dispensasi
- [ ] Statistik berdasarkan jenis status

## Bulk Attendance (Presensi Manual Kelas)

### Endpoint: `POST /admin/attendance/bulk`

**Request Body (JSON):**
```json
{
  "class_id": 1,
  "date": "2026-09-13",
  "students": [
    {"student_id": 1, "status": "Hadir"},
    {"student_id": 2, "status": "Terlambat"},
    {"student_id": 3, "status": "Sakit"},
    {"student_id": 4, "status": "Dispensasi"},
    {"student_id": 5, "status": "Alpha"}
  ]
}
```

**Response:**
```json
{
  "status": "success",
  "message": "5 kehadiran berhasil disimpan"
}
```

### Status yang Tersedia di Bulk
- Hadir
- Terlambat
- Sakit
- Dispensasi
- Alpha

### Catatan
- Siswa dengan status "Unknown" atau tidak dipilih akan di-skip
- Timestamp menggunakan tanggal yang dipilih + waktu saat ini
- Jika siswa sudah punya log di tanggal tersebut, data akan di-update (bukan di-insert ulang)
