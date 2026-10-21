# SMK NIBA Super Apps

Sistem informasi sekolah terpadu untuk SMK berbasis web — menggabungkan absensi RFID, portal siswa mobile-first, gamifikasi bahasa Inggris, sistem poin, notifikasi WhatsApp, dan manajemen bel otomatis dalam satu platform.

---

## Fitur Utama

### Learning Management System (LMS &amp; Ruang KBM)
- **Portal Guru (`/guru/login`, `/guru/dashboard`, `/guru/class/:id`)**:
  - Login NIP mandiri guru dengan manajemen sesi aman berbasis token DB &amp; cookie HTTP-only
  - Dashboard penugasan mengajar: daftar kelas &amp; mapel yang diampu dengan banner visual
  - Pembuatan &amp; pengelolaan sesi pertemuan KBM (urutan, judul, tanggal, deskripsi)
  - Distribusi bahan ajar: berkas dokumen (PDF, Word, PPT), video materi, dan tautan referensi eksternal
  - Presensi pertemuan PJJ (Pembelajaran Jarak Jauh) real-time dengan status Hadir, Izin, Sakit, Alpa
  - Penugasan daring: instruksi tugas, batas pengumpulan (*due date*), penerimaan berkas/link siswa, penilaian, serta feedback
  - Kuis online interaktif: kuis pilihan ganda dengan timer mundur, passing score, dan penilaian otomatis
- **Portal Siswa KBM (`/student/app` tab Belajar)**:
  - Ruang mata pelajaran semester aktif lengkap dengan banner ilustrasi
  - Akses materi ajar per sesi pertemuan
  - Status kehadiran KBM per pertemuan
  - Form penyerahan tugas (unggah dokumen / lampiran link URL)
  - Pengerjaan kuis digital langsung dari browser dengan hasil skor instan
- **Sistem Banner Vektor Otomatis (`public/assets/banners/`)**:
  - 8 Vektor SVG bawaan: Matematika (`math`), Teknologi/Informatika (`tech`), Bahasa (`language`), Sains/IPA (`science`), Bisnis/MPLB (`business`), Seni Budaya (`art`), Agama &amp; Budi Pekerti (`religion`), serta Literasi Umum (`general`)
  - Penugasan otomatis banner berbasis kategori &amp; nama mata pelajaran dengan fallback cerdas
  - Opsi pemilihan visual preset banner sekali klik dengan *live preview* di Admin Akademik (`views/admin/academics.html`)
- **Dukungan Cloud Storage Fleksibel**:
  - Dukungan penyimpanan berkas tugas &amp; materi di folder lokal (`public/uploads`) maupun Cloudflare R2 / AWS S3 kompatibel via `StorageService` (`internal/service/storage.go`)

### Absensi &amp; Presensi
- Check-in/Check-out otomatis via **RFID** dan **Face Recognition**
- Presensi manual per siswa maupun bulk per kelas
- Presensi sholat (Dzuhur & Ashar) dengan terminal scan terpisah
- Laporan harian, mingguan, bulanan — tampil di browser & export PDF
- Notifikasi WhatsApp real-time ke orang tua (via OneSender)

### Portal Siswa (Mobile-First SPA)
- **Modern FinTech Violet UI Redesign**: desain mobile-first modern bertema Electric Violet Gradient (`views/student/app.html`)
- Login NIS + PIN responsif (`views/student/login.html`) serta autentikasi **Reverse OTP &amp; Magic Link**
- **Ruang KBM / LMS Terintegrasi**: tab Belajar mandiri dengan materi per pertemuan, penugasan, dan kuis online
- **Multi-Child Account Switcher**: kemudahan bagi orang tua/wali dengan lebih dari satu anak untuk berpindah akun instan
- **Hero Attendance Card**: widget presensi gerbang realtime (Jam Masuk, Jam Pulang, live status badge)
- **Doa Harian &amp; Dzikir**: modul doa interaktif terintegrasi API `equran.id` (caching in-memory, pencarian, dan salin doa)
- Kartu Pelajar Digital dengan QR Code dinamis (auto-refresh 30 detik)
- Kalender kehadiran bulanan interaktif dengan visualisasi status
- Riwayat &amp; saldo poin (Poin Karakter / Panca &amp; Poin Prestasi)
- Widget cuaca lokal dan pengingat jadwal bel berikutnya
- Ganti PIN mandiri

### Manajemen Sarpras & Inventaris (KIB)
- Katalog aset berstandar KIB (KIB A s.d. E: Tanah, Bangunan, Mesin/Peralatan, dll.)
- Sirkulasi peminjaman barang (Siswa & Guru) dengan workflow persetujuan & pengembalian
- Manajemen tiket perawatan & pelaporan kerusakan barang
- Generator stiker label QR Code otomatis untuk setiap aset
- Export rekap inventaris ke spreadsheet Microsoft Excel (.xlsx)

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

### Halaman Error Sistem (404 &amp; 403)
- Redesain modern bertema **shadcn/ui (Zinc / Slate)** (`views/404.html` &amp; `views/403.html`)
- Dukungan penuh Dark Mode &amp; Light Mode dengan script inisialisasi instan tanpa kedipan (*zero FOUC*)
- Ikon Lucide (`compass`, `shield-alert`) dan tipografi mono modern
- Penanganan error terpusat (`e.HTTPErrorHandler`) dengan pemisahan otomatis: navigasi browser merender halaman responsif, sedangkan request endpoint `/api/*` merespons format JSON standar
- Rute preview langsung: [`/404`](http://localhost:8080/404) dan [`/403`](http://localhost:8080/403)

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
│   │   ├── asset_excel.go       # Sarpras Excel export
│   │   ├── asset_management.go  # Sarpras asset CRUD & QR
│   │   ├── attendance.go        # Attendance handlers
│   │   ├── audio.go             # Audio library
│   │   ├── auth.go              # Authentication
│   │   ├── doa.go               # Daily Prayer (equran.id proxy + cache)
│   │   ├── english.go           # English Daily Quest
│   │   ├── face.go              # Face recognition
│   │   ├── holiday.go           # Holiday management
│   │   ├── lms_admin.go         # Kurikulum mapel & banner assignment
│   │   ├── lms_materials_assignments.go # Bahan ajar & penugasan siswa
│   │   ├── lms_quiz.go          # Bank soal & evaluasi kuis online
│   │   ├── magic_auth.go        # Reverse OTP & Magic Link login
│   │   ├── parent.go            # Multi-child parent portal
│   │   ├── point_management.go  # Point & rules management
│   │   ├── report.go            # Reports (PDF & preview)
│   │   ├── sarpras_page.go      # Sarpras integrated page
│   │   ├── student.go           # Student CRUD
│   │   ├── student_portal.go    # Student portal API
│   │   ├── staff.go             # Staff CRUD
│   │   ├── teacher_kbm.go       # Pertemuan KBM guru & presensi PJJ
│   │   ├── teacher_portal.go    # Autentikasi NIP & profile guru
│   │   ├── wa.go / wa_handler.go # WhatsApp integration
│   │   └── ...
│   ├── models/
│   │   └── lms.go               # Struct & DTO modul LMS
│   ├── repository/
│   │   └── db.go                # DB init, migrations, seeds
│   ├── service/
│   │   └── storage.go           # File storage (Local & Cloudflare R2/S3)
│   ├── config/
│   │   ├── config.go            # Config loader
│   │   └── logging.go           # Structured logging
│   └── router/
│       └── router.go            # Route definitions & HTTPErrorHandler
├── pkg/
│   ├── pdf/                     # PDF generation helpers
│   ├── qrcode/                  # QR code helpers
│   ├── onesender/               # WhatsApp client
│   └── utils/                   # Utilities (phone format, file, photo)
├── public/
│   ├── assets/
│   │   ├── banners/             # 8 Vector SVG Banners (math, tech, dll)
│   │   ├── audio/               # Berkas nada bel MP3/WAV
│   │   └── photos/              # Pasfoto siswa & staf
│   └── uploads/                 # Berkas upload modul ajar & tugas siswa
├── views/
│   ├── admin/                   # Modular Admin UI (Shadcn SSR)
│   │   ├── layout_head.html     # Semantic tokens, Tailwind, CSRF setup
│   │   ├── layout_sidebar.html  # Clean sidebar navigation
│   │   ├── layout_topbar.html   # Sticky topbar & dark mode toggle
│   │   ├── layout_footer.html   # Clean minimalist footer
│   │   ├── dashboard.html       # Admin stat cards & live feed
│   │   ├── academics.html       # Master Siswa, Guru, Mapel, Plot KBM
│   │   ├── attendance.html      # Rekap & log kehadiran harian
│   │   ├── points.html          # Sistem poin karakter & prestasi
│   │   ├── sarpras.html         # Inventaris aset, pinjam, tiket
│   │   ├── schedules.html       # Pengaturan bel terjadwal
│   │   ├── settings.html        # Pengaturan sistem & sekolah
│   │   └── student_detail.html  # Detail & biodata siswa
│   ├── teacher/                 # Portal Guru KBM (Shadcn SSR)
│   │   ├── login.html           # Login NIP guru
│   │   ├── dashboard.html       # Daftar kelas diampu dengan kartu banner
│   │   └── class_detail.html    # Modul pertemuan, materi, tugas & kuis
│   ├── student/
│   │   ├── app.html             # Student portal SPA (Mobile-first violet)
│   │   └── login.html           # Student login modern
│   ├── 404.html                 # Error 404 Not Found (Shadcn modern)
│   ├── 403.html                 # Error 403 Access Denied (Shadcn modern)
│   ├── scan.html                # RFID scan terminal
│   ├── scan_prayer.html         # Prayer scan terminal
│   ├── idcard_single.html       # Digital PVC ID Card generator
│   └── index.html               # Public landing page
├── tests/                       # Modular test suite (LMS, auth, point, dll)
├── migrations/                  # SQL migration files
├── face_service/                # Python face recognition microservice
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
| GET | `/` | Landing page / dashboard ekosistem publik |
| GET | `/guru/login` | Halaman login guru |
| GET | `/student/login` | Halaman login portal siswa |
| GET | `/login` | Halaman login admin |
| GET | `/scan` | Terminal scan RFID gerbang |
| GET | `/scan-face` | Terminal kiosk presensi biometrik wajah |
| GET | `/scan-sholat` | Terminal scan sholat berjamaah |
| GET | `/404` | Halaman error 404 (Not Found) |
| GET | `/403` | Halaman error 403 (Access Denied) |
| GET | `/api/attendance/record` | Catat absensi RFID (`?rfid=`) |
| GET | `/api/sync` | Sync jadwal & pengumuman ke device |

### Teacher Portal (`/api/guru/*`, require teacher session)
| Method | Path | Deskripsi |
|--------|------|-----------|
| POST | `/api/guru/login` | Login guru (NIP + Password) |
| POST | `/api/guru/logout` | Logout sesi guru |
| GET | `/api/guru/profile` | Profil guru & daftar kelas diampu |
| PUT | `/api/guru/password` | Ubah password akun guru |
| GET | `/api/guru/meetings/:assignment_id` | Daftar sesi pertemuan KBM kelas |
| POST | `/api/guru/meetings` | Buat sesi pertemuan KBM baru |
| GET | `/api/guru/meetings/:id/attendance` | Rekap kehadiran siswa per pertemuan |
| POST | `/api/guru/meetings/:id/attendance` | Simpan presensi PJJ pertemuan KBM |
| POST | `/api/guru/materials` | Tambah materi ajar (dokumen/video/link) |
| DELETE | `/api/guru/materials/:id` | Hapus berkas/materi ajar |
| POST | `/api/guru/assignments` | Buat tugas baru dengan due date |
| GET | `/api/guru/assignments/:id/submissions` | Daftar pengumpulan tugas siswa |
| POST | `/api/guru/submissions/grade` | Beri nilai & feedback tugas siswa |
| POST | `/api/guru/quizzes` | Buat kuis online baru |
| GET | `/api/guru/quizzes/:id/questions` | Daftar soal kuis pilihan ganda |
| POST | `/api/guru/quizzes/questions` | Tambah butir pertanyaan kuis |
| DELETE | `/api/guru/quizzes/questions/:id` | Hapus butir pertanyaan kuis |
| DELETE | `/api/guru/quizzes/:id` | Hapus kuis online |
| GET | `/api/guru/quizzes/:id/attempts` | Rekap pengerjaan kuis siswa |

### Student Portal & LMS (`/api/student/*`, require student session)
| Method | Path | Deskripsi |
|--------|------|-----------|
| POST | `/api/student/login` | Login siswa (NIS + PIN / Reverse OTP) |
| GET | `/api/student/dashboard` | Data dashboard & hero attendance |
| GET | `/api/student/courses` | Daftar mapel semester aktif dengan banner |
| GET | `/api/student/courses/:id` | Detail KBM: materi, tugas & kuis per sesi |
| POST | `/api/student/assignments/submit` | Kumpul tugas siswa (upload berkas/link) |
| GET | `/api/student/quizzes/:id` | Ambil soal kuis untuk dikerjakan |
| POST | `/api/student/quizzes/submit` | Serahkan jawaban kuis & hitung skor |
| GET | `/api/student/calendar` | Kalender kehadiran bulanan |
| GET | `/api/student/qrcard` | Kartu QR digital dinamis |
| GET | `/api/student/profile` | Profil siswa |
| PUT | `/api/student/pin` | Ganti PIN mandiri |
| GET | `/api/student/points` | Saldo & riwayat poin karakter |
| GET | `/api/student/english/quest` | Quest harian bahasa Inggris |
| POST | `/api/student/english/submit` | Submit setoran quest harian |
| GET | `/api/student/doa` | Daftar & pencarian doa harian |

### Admin (`/admin/*`, require admin auth)
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/admin` | Dashboard admin & live feed |
| GET | `/admin/subjects` | Daftar master mata pelajaran & banner |
| POST | `/admin/subject/add` | Tambah mata pelajaran dengan banner |
| POST | `/admin/subject/update/:id` | Edit mata pelajaran & banner |
| DELETE | `/admin/subject/:id` | Hapus mata pelajaran |
| GET | `/admin/teacher-assignments` | Plot guru pengampu kelas & mapel |
| POST | `/admin/teacher-assignment/add` | Plot pengajar baru ke kelas |
| DELETE | `/admin/teacher-assignment/:id` | Hapus plot pengajar |
| POST | `/admin/student/add` | Tambah siswa baru |
| POST | `/admin/student/update/:id` | Edit biodata siswa |
| POST | `/admin/english/quest` | Buat quest manual |
| POST | `/admin/ai/generate-quest` | Generate quest otomatis dengan AI |
| POST | `/admin/ai/settings` | Simpan konfigurasi API AI |
| POST | `/admin/ai/test` | Test koneksi API AI |
| GET | `/admin/points/leaderboard` | Papan prestasi peringkat siswa |
| GET | `/admin/points/history` | Riwayat transaksi poin |
| GET | `/admin/recent-activity` | Feed aktivitas waktu-nyata |

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
