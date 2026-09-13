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
- Dua opsi:
  - **Dengan Surat Keterangan**: Checkbox "Dengan Surat Keterangan" dicentang
  - **Dengan Keterangan**: Textarea diisi tanpa centang checkbox
  - **Tanpa Keterangan**: Textarea kosong, checkbox tidak dicentang
- Status disimpan sebagai:
  - `Sakit (Surat)` jika ada surat
  - `Sakit (Keterangan)` jika ada keterangan teks
  - `Sakit` jika tanpa keterangan

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
has_letter: boolean (optional) - true/false, hanya untuk status Sakit
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
  - Has letter checkbox: tampil untuk Sakit

### Form Validation
- Status wajib dipilih
- Waktu wajib diisi untuk Hadir/Terlambat
- Keterangan wajib diisi untuk Dispensasi
- Keterangan opsional untuk Sakit

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
1. `views/admin.html` - UI modal dan JavaScript
2. `main.go` - Handler `ManualAttendanceHandler`
3. `internal/repository/db.go` - Migration untuk kolom `note`

## Testing Checklist
- [ ] Absen Hadir dengan jam custom
- [ ] Absen Terlambat dengan jam custom
- [ ] Sakit tanpa keterangan
- [ ] Sakit dengan keterangan teks
- [ ] Sakit dengan surat keterangan
- [ ] Dispensasi dengan keterangan (wajib)
- [ ] Alpha tanpa keterangan
- [ ] Validasi form (wajib waktu, wajib keterangan)
- [ ] WA notification untuk Hadir/Terlambat
- [ ] Data tersimpan dengan benar di DB

## Future Improvements
- [ ] History log untuk melihat keterangan
- [ ] Upload attachment untuk surat keterangan
- [ ] Approval workflow untuk dispensasi
- [ ] Statistik berdasarkan jenis status
