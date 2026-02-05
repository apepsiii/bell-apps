# SmartBell - Sistem Manajemen Sekolah

## Overview

**SmartBell** adalah sistem manajemen sekolah komprehensif yang dibangun dengan Go, menggabungkan:
1. **Sistem Bel Otomatis** - Manajemen jadwal bel sekolah dengan pemutaran audio otomatis
2. **Sistem Absensi RFID** - Tracking kehadiran siswa dan staff berbasis RFID
3. **Integrasi WhatsApp** - Notifikasi otomatis ke orang tua dan grup kelas
4. **Manajemen Data Akademik** - Data siswa, staff, kelas, dan jurusan
5. **Dashboard Admin** - Interface web lengkap untuk manajemen sistem

---

## Struktur Project

```
bell/
├── main.go                 # Aplikasi utama (1,892 baris) - Semua logika backend
├── go.mod                  # Dependensi Go module
├── go.sum                  # Checksum dependensi
├── database.db             # Database SQLite (dibuat saat runtime)
├── bell_linux              # Binary Linux hasil kompilasi
├── .air.toml              # Konfigurasi hot-reload untuk development
├── .gitignore             # Aturan Git ignore
├── README.md              # Instruksi setup dasar
├── deploy.md              # Panduan deployment VPS
├── setup.sh               # Script setup otomatis Linux
├── setup_nginx.sh         # Script konfigurasi Nginx + SSL
│
├── views/                 # Template HTML (1,902 total baris)
│   ├── admin.html         # Dashboard admin utama (933 baris)
│   ├── index.html         # Interface bell player (189 baris)
│   ├── login.html         # Halaman login admin (99 baris)
│   ├── scan.html          # Terminal absensi RFID (232 baris)
│   ├── profile.html       # Profil siswa/staff (208 baris)
│   └── admin_dashboard.html # Dashboard legacy (241 baris)
│
└── public/                # Asset statis
    ├── manifest.json      # PWA manifest
    ├── sw.js              # Service worker
    ├── index.html         # Landing page publik
    └── assets/
        ├── audio/         # File audio untuk bel
        │   └── .gitkeep
        └── photos/        # Foto siswa (kartu RFID)
            └── .gitkeep
```

---

## Teknologi & Framework

### Backend
- **Bahasa**: Go 1.25.6
- **Web Framework**: Echo v4 (Labstack)
- **Database**: SQLite dengan driver pure-Go ModernC
- **Template Engine**: Go's html/template

### Frontend
- **CSS Framework**: Tailwind CSS (via CDN)
- **Icons**: Lucide Icons
- **Charts**: Chart.js
- **Notifications**: SweetAlert2
- **PWA Support**: Service Worker + Manifest

### Integrasi Eksternal
- **WhatsApp API**: OneSender untuk notifikasi
- **Export Excel**: xuri/excelize untuk export data
- **PDF Generation**: gofpdf library

### Deployment
- **Systemd**: Manajemen service di Linux
- **Nginx**: Reverse proxy dengan SSL support
- **Certbot**: Manajemen sertifikat SSL otomatis

---

## Fitur Utama

### 1. **Sistem Penjadwalan Bel**
- Buat/edit/hapus jadwal bel dengan file audio custom
- Pemutaran otomatis di tablet/speaker sesuai jadwal
- Upload dan manajemen library audio
- Sinkronisasi jadwal real-time ke perangkat client

### 2. **Sistem Absensi RFID**
- **Siswa**: Check-in/check-out berbasis RFID dengan dukungan foto
- **Staff**: Tracking kehadiran terpisah untuk guru dan staff
- **Absensi Manual**: Admin dapat menandai kehadiran manual
- **Status Types**: Datang (Tepat Waktu), Terlambat, Pulang, Sakit, Izin, Alpha
- **Pencegahan Duplikat**: Mencegah multiple check-in per hari
- **Logika Berbasis Waktu**: Menentukan status otomatis berdasarkan window waktu yang dapat dikonfigurasi

### 3. **Notifikasi WhatsApp**
- **Notifikasi Orang Tua**: Pesan otomatis saat siswa datang/pulang
- **Pesan Grup Kelas**: Broadcast ke grup WhatsApp kelas
- **Notifikasi Staff**: Pesan personal untuk kehadiran guru
- **Template Customizable**: Variabel template seperti {name}, {time}, {status}
- **Dukungan Gambar**: Opsi mengirim dengan attachment foto

### 4. **Manajemen Data Akademik**
- **Jurusan**: Organisasi siswa berdasarkan jurusan/program
- **Kelas**: Manajemen kelas dengan integrasi grup WhatsApp
- **Siswa**: Profil lengkap dengan RFID, NIS, foto, kontak orang tua
- **Staff**: Profil guru dan staff dengan penunjukan peran
- **Import CSV**: Import bulk untuk siswa dan staff

### 5. **Dashboard & Analytics**
- **Statistik Real-time**: Total siswa, staff, jadwal
- **Chart Kehadiran**: 
  - Progress kelas mingguan (line chart)
  - Distribusi status (pie chart)
  - Trend rata-rata waktu kedatangan
- **Log Hari Ini**: Feed kehadiran live
- **List Hadir/Tidak Hadir**: List siswa yang dapat difilter
- **Calendar View**: Riwayat kehadiran bulanan per siswa/staff

### 6. **Manajemen Perangkat**
- Register perangkat tablet/speaker
- Monitor status online/offline
- Track waktu sinkronisasi terakhir
- API endpoint untuk perangkat IoT

---

## File Penting

### File Inti

**main.go** (1,892 baris)
- Inisialisasi database dan migrasi
- Semua HTTP handler dan business logic
- Operasi CRUD untuk semua entitas
- Logika pencatatan kehadiran
- Fungsi integrasi WhatsApp
- Generasi data chart
- Fungsionalitas export CSV
- Middleware autentikasi

**go.mod**
```go
module belsekolah
go 1.25.6

require (
    github.com/labstack/echo/v4 v4.15.0
    modernc.org/sqlite v1.44.3
)
```

### File Konfigurasi

**.air.toml**
- Konfigurasi hot-reload untuk development
- Auto-rebuild saat perubahan file .go, .html
- Exclude database dan asset publik

**.gitignore**
- Ignore binaries (.exe, .dll)
- Ignore file database
- Ignore audio/foto yang diupload (menjaga struktur via .gitkeep)

### Script Deployment

**setup.sh** (85 baris)
- Wizard instalasi VPS otomatis
- Set file permissions
- Buat systemd service
- Konfigurasi port settings
- Start dan enable service

**setup_nginx.sh** (84 baris)
- Install Nginx dan Certbot
- Buat konfigurasi reverse proxy
- Enable SSL dengan Let's Encrypt
- Dukungan custom domain

### Views (Templates)

**views/admin.html** (933 baris)
- Dashboard admin single-page
- 7 section utama: Dashboard, Schedules, Audio, Majors, Classes, Students, Attendance
- Sinkronisasi jam real-time
- Operasi CRUD berbasis AJAX
- Navigasi sidebar responsive
- Visualisasi Chart.js

**views/index.html** (189 baris)
- Interface bell player untuk tablet
- Auto-play audio sesuai jadwal
- Tampilan jam besar
- Countdown bel berikutnya
- Sinkronisasi jadwal setiap 5 menit
- UI cantik dengan Tailwind CSS

**views/scan.html** (232 baris)
- Interface terminal absensi RFID
- Auto-focus input untuk RFID reader
- State visual: Idle → Loading → Success/Error
- Tampilkan foto siswa saat sukses
- Auto-reset setelah 3 detik
- Mode kiosk fullscreen

**views/login.html** (99 baris)
- Halaman autentikasi admin
- Desain split-screen dengan branding
- Validasi form
- Handling pesan error
- Kredensial default: admin/admin123

**views/profile.html** (208 baris)
- Halaman detail siswa/staff
- Tampilan kalender bulanan
- Riwayat kehadiran
- Tampilan QR code
- Tampilan foto

---

## Skema Database

**11 Tabel Utama:**

1. **schedules** - Entri jadwal bel (time, label, audio_file)
2. **audio_files** - Library audio (file_name, display_name)
3. **devices** - Tablet/speaker yang terdaftar
4. **majors** - Jurusan/program sekolah
5. **classes** - Kelas dengan asosiasi jurusan dan WhatsApp group ID
6. **students** - Data siswa dengan RFID, NIS, foto, nomor telepon orang tua
7. **staff** - Data guru/staff dengan RFID, NIP, role
8. **attendance_logs** - Semua record kehadiran dengan timestamp, status, method (RFID/MANUAL)
9. **attendance_settings** - Konfigurasi key-value untuk window waktu dan template WhatsApp
10. **running_texts** - Konten running text digital signage
11. **signage_media** - File media digital signage

---

## API Endpoints

### Public Endpoints
```
GET  /                          # Halaman bell player
GET  /login                     # Halaman login admin
GET  /scan                      # Halaman terminal RFID
POST /api/login                 # Autentikasi
POST /api/logout                # Logout
GET  /api/sync                  # Sinkronisasi jadwal untuk perangkat
GET  /api/attendance/record     # Recording kehadiran RFID (endpoint IoT)
```

### Admin Endpoints (Memerlukan Autentikasi)
```
GET  /admin                     # Dashboard

# Schedules
POST   /admin/schedule/add
POST   /admin/schedule/update/:id
DELETE /admin/schedule/:id

# Audio
POST   /admin/audio/upload
POST   /admin/audio/rename/:id
DELETE /admin/audio/:id

# Devices
POST   /admin/device/add
POST   /admin/device/update/:id
DELETE /admin/device/:id

# Majors
POST   /admin/major/add
POST   /admin/major/update/:id
DELETE /admin/major/:id

# Classes
POST   /admin/class/add
POST   /admin/class/update/:id
DELETE /admin/class/:id

# Students
POST   /admin/student/add
POST   /admin/student/update/:id
DELETE /admin/student/:id
POST   /admin/student/import      # Import bulk CSV

# Staff
POST   /admin/staff/add
POST   /admin/staff/update/:id
DELETE /admin/staff/:id
POST   /admin/staff/import         # Import bulk CSV

# Attendance
POST /admin/attendance/settings    # Update window waktu & template WhatsApp
POST /admin/attendance/manual      # Penandaan kehadiran manual
POST /admin/attendance/test-wa     # Test integrasi WhatsApp
GET  /admin/student/calendar       # Kehadiran bulanan siswa
GET  /admin/staff/calendar         # Kehadiran bulanan staff

# Profiles
GET  /admin/student/:id            # Halaman profil siswa
GET  /admin/staff/:id              # Halaman profil staff
```

---

## Konfigurasi Default

**Kredensial Admin:**
- Username: `admin`
- Password: `admin123`

**Port Default:** 8080 (dapat dikonfigurasi via environment variable PORT)

**Window Waktu Kehadiran:**
- Kedatangan: 06:00 - 07:15
- Kepulangan: 15:30 - 17:00

---

## Opsi Deployment

### Opsi 1: Binary Deployment
1. Compile untuk Linux: `GOOS=linux GOARCH=amd64 go build -o bell_linux main.go`
2. Upload: `bell_linux`, `views/`, `public/` ke VPS
3. Jalankan script setup: `bash setup.sh`

### Opsi 2: Git Deployment
1. Push code ke GitHub
2. Clone di VPS
3. Install Go: `sudo apt install golang`
4. Jalankan: `go run main.go`

### Production Setup
1. Jalankan `bash setup.sh` untuk membuat systemd service
2. Jalankan `sudo bash setup_nginx.sh` untuk konfigurasi domain dan SSL
3. Akses via HTTPS dengan custom domain

---

## Tujuan Project

SmartBell dirancang sebagai **solusi manajemen sekolah all-in-one** untuk sekolah Indonesia (SMK/SMA), menggabungkan:

1. **Otomasi** - Eliminasi penekanan bel manual dengan pemutaran audio terjadwal
2. **Akuntabilitas** - Tracking kehadiran berbasis RFID dengan verifikasi foto
3. **Komunikasi** - Notifikasi WhatsApp real-time ke orang tua dan staff
4. **Manajemen Data** - Data akademik terpusat dengan kemampuan export
5. **Affordability** - Menggunakan SQLite (tidak perlu database server), dapat di-deploy di VPS murah

Sistem ini khususnya cocok untuk sekolah kejuruan (SMK) yang perlu melacak kehadiran siswa dan guru, mengelola berbagai jurusan/kelas, dan menjaga komunikasi dengan orang tua.

---

## Catatan Teknis

- **Monolithic Application**: Semua logika dalam satu file `main.go` untuk kesederhanaan
- **Lightweight**: Go + Echo + SQLite membuatnya ringan dan mudah di-deploy
- **Production-Ready**: Error handling komprehensif dan otomasi deployment
- **PWA Support**: Dapat diinstall sebagai aplikasi di tablet/smartphone
- **Indonesian Language**: UI dan pesan dalam Bahasa Indonesia untuk sekolah lokal

---

## Development

### Prerequisites
- Go 1.25.6 atau lebih baru
- Air (untuk hot-reload development)

### Running Locally
```bash
# Install dependencies
go mod download

# Run with hot-reload
air

# Or run directly
go run main.go
```

### Building
```bash
# Build untuk Linux
GOOS=linux GOARCH=amd64 go build -o bell_linux main.go

# Build untuk Windows
GOOS=windows GOARCH=amd64 go build -o bell.exe main.go
```

---

## Kontributor

Project ini adalah sistem manajemen sekolah yang dirancang khusus untuk SMK/SMA di Indonesia dengan fokus pada kemudahan penggunaan dan affordability.

---

**Last Updated:** February 2026
