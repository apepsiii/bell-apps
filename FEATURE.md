# FEATURE.md — SMK NIBA Super Apps

Dokumen ini merangkum seluruh fitur yang tersedia pada **Halaman Admin** dan **Frontend Siswa (Portal Siswa)**. Digunakan sebagai dasar penulisan unit test dan checklist fungsionalitas.

---

## A. HALAMAN ADMIN

Halaman admin di-render oleh `app.DashboardHandler` pada route `GET /admin` (file `views/admin.html`). Menggunakan navigasi hash-based dengan 6 dropdown menu utama.

### A1. Dashboard (Ringkasan)
- Jam server real-time.
- Kartu statistik: Siswa Aktif, Siswa Tidak Aktif, Total Guru/Staff, Bel Berikutnya.
- 3 Chart (Chart.js):
  - Progress Presensi Mingguan per Kelas (line).
  - Sebaran Status Mingguan (doughnut).
  - Tren Jam Kedatangan Mingguan (line, sumbu Y format HH:MM).
- Papan Poin Siswa (leaderboard preview).
- **Route:** `GET /admin`

### A2. Manajemen Jadwal (Bell Schedule)
- CRUD jadwal bel: Waktu, Keterangan, Audio.
- Preview audio per item.
- **Routes:**
  - `POST /admin/schedule/add`
  - `POST /admin/schedule/update/:id`
  - `DELETE /admin/schedule/:id`

### A3. Audio Library
- Upload (`.mp3/.wav/.ogg`), preview, rename, delete.
- Progress bar upload.
- **Routes:**
  - `POST /admin/audio/upload`
  - `POST /admin/audio/rename/:id`
  - `DELETE /admin/audio/:id`

### A4. Data Jurusan (Majors)
- CRUD jurusan sekolah.
- **Routes:**
  - `POST /admin/major/add`
  - `POST /admin/major/update/:id`
  - `DELETE /admin/major/:id`

### A5. Data Kelas (Classes)
- CRUD kelas, terhubung ke jurusan.
- Menyimpan WhatsApp Group ID untuk broadcast WA.
- **Routes:**
  - `POST /admin/class/add`
  - `POST /admin/class/update/:id`
  - `DELETE /admin/class/:id`
  - `GET /admin/classes/json`

### A6. Data Siswa (Students)
- CRUD siswa lengkap dengan tabel searchable/sortable/filterable.
- Field: RFID, NIS, NIS Siswa, Kelas, Nama, Nama Ortu, No HP Ortu, Tgl Lahir, Status, Password, Foto (upload/kamera).
- Bulk: Aktifkan, Nonaktifkan, Naik Kelas, Hapus.
- Import CSV & JSON.
- Kartu Pelajar (ID Card) dengan QR.
- Profil siswa terpisah dengan kalender kehadiran.
- **Routes:**
  - `POST /admin/student/add`
  - `POST /admin/student/update/:id`
  - `POST /admin/student/status/:id`
  - `DELETE /admin/student/:id`
  - `POST /admin/student/import`
  - `POST /admin/student/import-json`
  - `POST /admin/students/delete-multiple`
  - `POST /admin/students/bulk-status`
  - `POST /admin/students/promote`
  - `GET /admin/students/json`
  - `GET /admin/student/:id`
  - `GET /admin/student/calendar`

### A7. Kartu Siswa (ID Cards)
- Generator kartu pelajar printable dengan QR code.
- Filter per kelas.
- **Routes:**
  - `GET /admin/idcard`
  - `GET /admin/student/idcard/:id`
  - `GET /admin/students/idcard`

### A8. Guru & Staff (Staff Management)
- CRUD guru/staff: NIP, Nama, Role, RFID, No HP.
- Import CSV.
- Profil staff dengan kalender kehadiran.
- **Routes:**
  - `POST /admin/staff/add`
  - `POST /admin/staff/update/:id`
  - `DELETE /admin/staff/:id`
  - `POST /admin/staff/import`
  - `GET /admin/staff/:id`
  - `GET /admin/staff/calendar`

### A9. Kontrol Presensi Siswa
- Dashboard real-time: Sudah Presensi vs Belum Datang.
- Presensi Manual per siswa (Hadir/Sakit/Alpha).
- Presensi Bulk per kelas (Hadir/Sakit/Izin/Alpha).
- Buka Terminal Scan (`/scan`).
- Pengaturan Jam.
- **Routes:**
  - `POST /admin/attendance/manual`
  - `GET /admin/attendance/daily`
  - `POST /admin/attendance/bulk`
  - `POST /admin/attendance/settings`

### A10. Laporan Kehadiran (PDF + WhatsApp)
- Filter: Tipe (Harian/Mingguan/Bulanan), Subjek (Siswa/Staff), Kelas, Tanggal.
- Preview di halaman + Download PDF.
- **Salin Laporan WA**: format teks WhatsApp dengan daftar hadir (Nama : Jam Datang) & tidak hadir (Sakit/Izin/Alpha).
- **Opsi Laporan WA**: modal dengan textarea + copy to clipboard.
- **Routes:**
  - `GET /admin/report/daily`
  - `GET /admin/report/weekly`
  - `GET /admin/report/monthly`

### A11. Manajemen Pengumuman (Announcements / TTS)
- Buat pengumuman teks → convert ke speech/audio.
- Penjadwalan opsional.
- Play Now, Delete.
- **Routes:**
  - `GET /admin/announcements`
  - `POST /admin/announcement/add`
  - `DELETE /admin/announcement/:id`
  - `POST /admin/announcement/play/:id`

### A12. English Daily Quest (Gamifikasi)
- 4 Tab: Buat Quest, Review, Leaderboard, Progress.
- Tipe Quest: Written, Vocabulary, Quiz, Voice.
- Review submission: Approve/Reject + feedback.
- **Routes:**
  - `GET /admin/english/quest`
  - `POST /admin/english/quest`
  - `GET /admin/english/submissions`
  - `POST /admin/english/submissions/:id/review`
  - `GET /admin/english/leaderboard`
  - `GET /admin/english/progress`

### A13. Sistem Presensi Sholat
- 3 Tab: Presensi Manual, Laporan Bulanan, Log Harian.
- Dzuhur/Ashar/PMS tracking per siswa.
- **Routes:**
  - `GET /admin/prayer/attendance`
  - `POST /admin/prayer/attendance`
  - `GET /admin/prayer/report`
  - `GET /api/attendance/prayer-logs`

### A14. Sistem Poin Siswa
- 4 Tab: Leaderboard, Input Poin, Tukar Poin (Redeem), Master Data.
- Master Data: Aturan (Prestasi/Pelanggaran) & Hadiah (Rewards).
- Point Claims dengan evidence (submission/approval workflow).
- **Routes:**
  - `GET /admin/point-rules`, `GET /admin/point-rules-v2`
  - `POST /admin/point-rules/add`, `DELETE /admin/point-rules/:id`
  - `GET /admin/point-rewards`, `POST /admin/point-rewards/add`, `DELETE /admin/point-rewards/:id`
  - `POST /admin/points/transaction`, `POST /admin/points/redeem`
  - `GET /admin/points/student/:id`, `GET /admin/points/leaderboard`
  - `POST /admin/point-claims`, `GET /admin/point-claims`
  - `POST /admin/point-claims/:id/approve`, `POST /admin/point-claims/:id/reject`

### A15. Hari Libur (Holidays)
- CRUD hari libur (Nasional/Internal).
- Import Nasional (standar).
- **Routes:**
  - `GET /admin/holidays`
  - `POST /admin/holiday/add`
  - `PUT /admin/holiday/:id`
  - `DELETE /admin/holiday/:id`
  - `POST /admin/holidays/import-national`

### A16. Perangkat (Devices)
- CRUD perangkat bel/scan: Nama, IP, Status.
- **Routes:**
  - `POST /admin/device/add`
  - `POST /admin/device/update/:id`
  - `DELETE /admin/device/:id`

### A17. Log Pengiriman WhatsApp
- Tabel log WA: Waktu, Tujuan, Pesan, Status, Response.
- Refresh via AJAX.
- **Route:** `GET /admin/wa-logs`

### A18. Pengaturan Sistem Presensi & WhatsApp
- Waktu Absensi (Mulai/Batas Datang & Pulang).
- Waktu Sholat (Dzuhur/Ashar).
- Integrasi WhatsApp OneSender (API URL, Token, 6 template: Hadir/Terlambat/Pulang/Staff Masuk/Staff Pulang/Ulang Tahun).
- Test Koneksi WA.
- **Routes:**
  - `POST /admin/attendance/settings`
  - `POST /api/attendance/test-wa`

### A19. Pengaturan Sekolah
- Konfigurasi hari kerja aktif (Senin–Minggu).
- **Routes:**
  - `GET /admin/settings/school`
  - `PUT /admin/settings/school`

### A20. Pengenalan Wajah (Face Recognition)
- Halaman registrasi wajah (dari profil siswa/staff).
- API: register, status, list, verify, delete.
- **Routes:**
  - `GET /admin/face/register/:id`
  - `POST /admin/face/register`
  - `GET /admin/face/status`
  - `GET /admin/faces`
  - `DELETE /admin/face/:student_id`
  - `POST /admin/face/verify`

### A21. QR Code Generation (Utility)
- Generate QR code base64 PNG dari RFID UID.
- **Route:** `GET /admin/qr-generate`

---

## B. FRONTEND SISWA (PORTAL SISWA)

Halaman siswa adalah SPA mobile-first (`views/student/app.html`) dengan 8 view section, login page, dan 5 tab bottom navigation.

### B1. Login Siswa
- Autentikasi NIS + PIN (4-8 digit).
- Default PIN `123456` jika password kosong.
- Session cookie 7 hari.
- **Routes:**
  - `GET /student/login`
  - `POST /api/student/login`

### B2. Logout Siswa
- Hapus session (cookie + DB).
- **Route:** `POST /api/student/logout`

### B3. Session & Middleware Auth
- Guard semua halaman & API siswa.
- **Middleware:** `handler.StudentAuth(db)`

### B4. Dashboard / Beranda (Home)
- Endpoint agregat tunggal: profil, kehadiran hari ini & kemarin, total poin, statistik bulanan, bel berikutnya, pengumuman terbaru.
- **Sub-fitur:**
  - **B4a.** Profile Header Card (foto, greeting "Assalamualaikum {2 kata pertama}", kelas, ikon piala & notifikasi).
  - **B4b.** Points Balance Card (total poin + toggle hide/show).
  - **B4c.** Announcement Banner Slider (auto-rotate 4 detik, gradient berubah).
  - **B4d.** Quick Action Menu (Presensi aktif; Jadwal/Nilai/E-Kantin disabled; Lainnya popup).
  - **B4e.** Kehadiran Hari Ini (Jam Masuk & Pulang, status-aware: Hadir/Terlambat/Sakit/Izin).
  - **B4f.** Attendance Insight (pesan motivasi dinamis: early/ontime/late/sick/permission, perbandingan dengan kemarin).
  - **B4g.** Riwayat Poin Terakhir (5 item terbaru).
  - **B4h.** Default PIN Reminder Toast.
- **Route:** `GET /api/student/dashboard`

### B5. Presensi / Riwayat Kehadiran (Calendar)
- Kalender bulanan dengan status berwarna per hari.
- Navigasi prev/next month.
- Detail per hari: status, jam masuk, jam pulang.
- Legend 5 status (Hadir/Terlambat/Izin/Sakit/Alpa).
- **Route:** `GET /api/student/calendar`

### B6. Kartu Pelajar Digital / QR Siswa
- Kartu digital dengan foto, nama, NIS, QR code.
- Animasi scan line.
- Download & Share (Web Share API).
- Auto-refresh 30 detik.
- **Route:** `GET /api/student/qrcard`

### B7. Profil Siswa
- Foto, nama, kelas.
- Info: NIS, Tanggal Lahir, Nama Wali, HP Wali.
- **B7a. Ganti PIN** (PIN Lama, PIN Baru, Konfirmasi; validasi 4-8 digit numerik).
- **B7b. Logout** button.
- **Routes:**
  - `GET /api/student/profile`
  - `PUT /api/student/pin`

### B8. Poin & Reward
- Kartu saldo poin (glassmorphism).
- Riwayat poin lengkap (up/down arrow, deskripsi, tanggal, nilai).
- **Route:** `GET /api/student/points`

### B9. Notifikasi
- List notifikasi dari pengumuman.
- Tombol "Tandai Baca" (placeholder).
- **Data:** dari dashboard `announcements`.

### B10. Pusat Informasi / Pengumuman
- List pengumuman terbaru.
- Empty state.
- **Data:** dari dashboard `announcements`.

### B11. English Daily Quest
- **B11a.** Stats Header (Streak, Total XP, Badge).
- **B11b.** Quest Hari Ini (4 tipe: Written, Vocabulary, Quiz, Voice).
- **B11c.** Submit Setoran Modal.
- **B11d.** Badge Kamu (badge yang diraih).
- **B11e.** Riwayat Setoran (30 terakhir, status + feedback guru).
- **B11f.** Leaderboard English (modal, ranking dengan medali).
- **Routes:**
  - `GET /api/student/english/quest`
  - `POST /api/student/english/submit`
  - `GET /api/student/english/profile`
  - `GET /api/student/english/leaderboard`

---

## Status Fitur (Catatan)

| Kategori | Status |
|----------|--------|
| Signage / Running Text | Tabel DB ada, **tidak ada route/UI admin** |
| Operator Management (admin-side) | Tidak ada UI admin (hanya self-service operator) |
| Quick Menu: Jadwal, Nilai, E-Kantin | **Disabled/Coming soon** (stub) |
| Notifikasi: Tandai Baca | **Placeholder** (alert, belum persisten) |
