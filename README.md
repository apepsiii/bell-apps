# SMK NIBA Super Apps

Sistem informasi sekolah terpadu untuk SMK berbasis web — menggabungkan absensi RFID, portal siswa mobile-first, gamifikasi bahasa Inggris, sistem poin, notifikasi WhatsApp, dan manajemen bel otomatis dalam satu platform.

---

## Fitur Utama

### Absensi & Presensi
- Check-in/Check-out otomatis via **RFID** dan **Face Recognition**
- Presensi manual per siswa maupun bulk per kelas
- Presensi sholat (Dzuhur & Ashar) dengan terminal scan terpisah
- Laporan harian, mingguan, bulanan — tampil di browser & export PDF
- Notifikasi WhatsApp real-time ke orang tua (via OneSender)

### Portal Siswa (Mobile-First SPA)
- Login NIS + PIN, session 7 hari
- Dashboard: rekap kehadiran, saldo poin, bel berikutnya, pengumuman
- Kartu Pelajar Digital dengan QR Code (auto-refresh 30 detik)
- Kalender kehadiran bulanan dengan status berwarna
- Riwayat & saldo poin lengkap
- Ganti PIN mandiri

### English Daily Quest (Gamifikasi)
- Quest harian: Written, Vocabulary, Quiz, Voice
- **AI Quest Generator** — generate quest otomatis via API OpenAI-compatible
- Submit setoran & review oleh guru (Approve/Reject + feedback)
- Sistem streak, XP, dan badge otomatis
- Leaderboard English

### Sistem Poin Siswa
- Input poin prestasi & pelanggaran via aturan yang bisa dikustomisasi
- Import aturan massal via CSV
- Penukaran poin dengan reward (stok terkontrol)
- Point Claims: siswa submit klaim, admin approve/reject
- Riwayat transaksi lengkap dengan filter
- Leaderboard poin terintegrasi dengan English XP

### Manajemen Sekolah
- CRUD Siswa (foto, RFID, NIS, data orang tua, status aktif/nonaktif)
- Import siswa via CSV & JSON, bulk naik kelas
- CRUD Guru/Staff, import CSV
- Manajemen Kelas & Jurusan, WhatsApp Group ID per kelas
- Kartu Pelajar (ID Card) printable dengan QR

### Bell & Audio
- Jadwal bel dengan audio custom (MP3/WAV/OGG)
- Library audio: upload, rename, preview, delete
- Pengumuman TTS (teks ke suara) dengan penjadwalan

### WhatsApp Integration
- Integrasi OneSender: 6 template pesan (Hadir, Terlambat, Pulang, Staff Masuk, Staff Pulang, Ulang Tahun)
- Auto-broadcast birthday greeting
- Log pengiriman WA lengkap
- Test koneksi WA dari dashboard

### Pengaturan & Konfigurasi
- Jam absensi masuk & pulang
- Hari kerja aktif (Senin–Minggu)
- Hari libur: CRUD + import nasional
- **Pengaturan AI**: Base URL, API Key, Model, Test Koneksi
- Pengenalan wajah (registrasi & verifikasi)

### Dashboard Admin
- Statistik real-time: siswa aktif/nonaktif, total staff, bel berikutnya
- Grafik Chart.js: progress presensi mingguan, sebaran status, tren jam kedatangan
- Papan Poin Siswa (leaderboard preview)
- **Aktivitas Terbaru**: feed real-time kehadiran, transaksi poin, English, point claims, pengumuman

---

## Tech Stack

| Layer | Teknologi |
|-------|-----------|
| Backend | Go 1.25+ + Echo v4 |
| Database | SQLite (default) / MySQL |
| Frontend | HTML + Tailwind CSS + Vanilla JS (embedded templates) |
| PDF | gofpdf |
| QR Code | go-qrcode |
| WhatsApp | OneSender API |
| Face Recognition | Python + face_recognition (microservice terpisah) |
| AI Quest | OpenAI-compatible REST API |

---

## Struktur Proyek

```
bell-apps/
├── main.go                      # Entry point & App struct
├── config.yaml                  # Konfigurasi server & database
├── internal/
│   ├── handler/                 # HTTP handlers
│   │   ├── activity.go          # Recent activity feed
│   │   ├── ai_quest.go          # AI quest generation
│   │   ├── ai_settings.go       # AI settings CRUD + test
│   │   ├── announcement.go      # Announcements
│   │   ├── attendance.go        # Attendance handlers
│   │   ├── audio.go             # Audio library
│   │   ├── auth.go              # Authentication
│   │   ├── english.go           # English Daily Quest
│   │   ├── face.go              # Face recognition
│   │   ├── holiday.go           # Holiday management
│   │   ├── point.go             # Point system
│   │   ├── point_history.go     # Point transaction history
│   │   ├── report.go            # Reports (PDF)
│   │   ├── student.go           # Student CRUD
│   │   ├── student_portal.go    # Student portal API
│   │   ├── staff.go             # Staff CRUD
│   │   ├── wa.go / wa_handler.go # WhatsApp integration
│   │   └── ...
│   ├── repository/
│   │   └── db.go                # DB init, migrations, seeds
│   ├── config/
│   │   ├── config.go            # Config loader
│   │   └── logging.go           # Structured logging
│   └── router/
│       └── router.go            # Route definitions
├── pkg/
│   ├── pdf/                     # PDF generation helpers
│   ├── qrcode/                  # QR code helpers
│   ├── onesender/               # WhatsApp client
│   └── utils/                   # Utilities (phone format, file, photo)
├── views/
│   ├── admin.html               # Admin SPA (hash-based navigation)
│   ├── student/
│   │   ├── app.html             # Student portal SPA (mobile-first)
│   │   └── login.html           # Student login
│   ├── scan.html                # RFID scan terminal
│   ├── scan_prayer.html         # Prayer scan terminal
│   ├── idcard.html              # ID Card generator
│   └── index.html               # Public leaderboard
├── migrations/                  # SQL migration files
├── face_service/                # Python face recognition microservice
│   ├── main.py
│   └── requirements.txt
├── *_test.go                    # Integration tests (package main)
└── FEATURE.md                   # Feature documentation
```

---

## Instalasi & Menjalankan

### Prasyarat
- Go 1.21+
- Python 3.8+ (hanya jika menggunakan face recognition)

### 1. Clone & Build

```bash
git clone https://github.com/apepsiii/bell-apps.git
cd bell-apps

# Build
go build -o bell-apps .

# Jalankan
./bell-apps
```

Server berjalan di `http://localhost:8080` secara default.

### 2. Build untuk deployment

```bash
# Linux/ARM64 (Raspberry Pi, server ARM)
GOOS=linux GOARCH=arm64 go build -o bell-apps-arm64 .

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o bell-apps-linux .
```

### 3. Face Service (opsional)

```bash
cd face_service
pip install -r requirements.txt
python main.py
```

---

## Konfigurasi

Saat pertama kali dijalankan, `config.yaml` akan dibuat otomatis. Edit sesuai kebutuhan:

```yaml
server:
  host: "0.0.0.0"
  port: "8080"

database:
  driver: "sqlite"   # atau "mysql"
  path: "data.db"

admin:
  username: "admin"
  password: "admin123"
```

Untuk MySQL:

```yaml
database:
  driver: "mysql"
  host: "localhost"
  port: "3306"
  name: "bellsekolah"
  user: "root"
  password: "yourpassword"
```

---

## API Endpoints

### Public
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/` | Landing page / leaderboard publik |
| GET | `/scan` | Terminal scan RFID |
| GET | `/scan-sholat` | Terminal scan sholat |
| GET | `/student/login` | Halaman login siswa |
| GET | `/api/attendance/record` | Catat absensi RFID (`?rfid=`) |
| GET | `/api/sync` | Sync jadwal & pengumuman ke device |

### Student Portal (`/api/student/*`)
| Method | Path | Deskripsi |
|--------|------|-----------|
| POST | `/api/student/login` | Login siswa |
| GET | `/api/student/dashboard` | Data dashboard siswa |
| GET | `/api/student/calendar` | Kalender kehadiran |
| GET | `/api/student/qrcard` | Kartu QR digital |
| GET | `/api/student/profile` | Profil siswa |
| PUT | `/api/student/pin` | Ganti PIN |
| GET | `/api/student/points` | Saldo & riwayat poin |
| GET | `/api/student/english/quest` | Quest hari ini |
| POST | `/api/student/english/submit` | Submit setoran |

### Admin (`/admin/*`, require auth)
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/admin` | Dashboard admin |
| POST | `/admin/student/add` | Tambah siswa |
| POST | `/admin/student/update/:id` | Edit siswa |
| POST | `/admin/english/quest` | Buat quest |
| POST | `/admin/ai/generate-quest` | Generate quest dengan AI |
| POST | `/admin/ai/settings` | Simpan pengaturan AI |
| POST | `/admin/ai/test` | Test koneksi AI |
| GET | `/admin/points/leaderboard` | Leaderboard poin |
| GET | `/admin/points/history` | Riwayat transaksi poin |
| GET | `/admin/recent-activity` | Feed aktivitas terbaru |

---

## Pengaturan AI Quest Generator

1. Buka **Admin → Pengaturan Sistem → Pengaturan AI**
2. Isi:
   - **Base URL**: endpoint OpenAI-compatible, contoh `https://api.openai.com/v1/chat/completions`
   - **API Key**: API key provider
   - **Model**: nama model, contoh `gpt-4o-mini`
3. Klik **Test Koneksi** untuk verifikasi
4. Buka **English Daily Quest → Buat Quest**, isi Topik, klik **Generate dengan AI**

Mendukung provider: OpenAI, Groq, Together AI, Ollama (lokal), dan provider lain yang kompatibel dengan format OpenAI.

---

## Menjalankan Tests

```bash
go test ./... -v
```

---

## Lisensi

MIT License — bebas digunakan dan dimodifikasi.
