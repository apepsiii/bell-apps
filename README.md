# SmartBell

Sistem manajemen sekolah lengkap untuk SMK dengan fitur absensi RFID, face recognition, dan bell otomatisasi.

## Fitur

### Manajemen Akademik
- **Siswa**: CRUD dengan foto, RFID, NIS, data orang tua
- **Guru/Staff**: Manajemen guru dan tenaga kependidikan
- **Kelas & Jurusan**: Pengaturan kelas dan jurusan
- **Promosi Siswa**: Pindah kelas massal

### Absensi
- **RFID**: Check-in/Check-out otomatis via kartu RFID
- **Manual**: Input absensi manual oleh admin
- **Face Recognition**: Verifikasi wajah untuk absensi (via face_service)
- **Sholat**: Pencatatan kehadiran siswa dalam sholat berjamaah

### Bell Otomatis
- **Jadwal Bell**: Atur jadwal bel dengan audio custom
- **Pengumuman**: Jadwalkan pengumuman dengan audio
- **Running Text**: Teks berjalan untuk display signage
- **Media Signage**: Tampilkan gambar/video di display

### Notifikasi & Komunikasi
- **WhatsApp (OneSender)**: Kirim notifikasi absensi ke orang tua
- **Grup WA Kelas**: Broadcast ke grup WhatsApp kelas
- **Template Pesan**: Customizable message templates
- **Ucapan Ulang Tahun**: Auto WhatsApp birthday greeting

### Sistem Poin
- **Poin Prestasi**: Tambah poin untuk prestasi siswa
- **Poin Pelanggaran**: Catat pelanggaran dan poin negatif
- **Voucher/Reward**: Tukar poin dengan voucher
- **Leaderboard**: Ranking siswa berdasarkan poin

### Pelaporan
- **Export CSV**: Export data absensi harian/bulanan/custom
- **Kalender Absensi**: View kalender per siswa/staff
- **Statistik**: Grafik kehadiran mingguan

### IoT & Device
- **Device Sync**: Sinkronisasi dengan device bell pintar
- **API RFID**: Endpoint untuk reader RFID eksternal
- **Dashboard Real-time**: Monitor kehadiran langsung

## Tech Stack

- **Backend**: Go 1.25+ dengan Echo framework
- **Database**: SQLite
- **Face Service**: Python dengan face_recognition library
- **Frontend**: HTML/JavaScript (embedded templates)
- **Notifications**: OneSender WhatsApp API

## Struktur Proyek

```
├── cmd/                  # CLI tools
├── internal/             # Internal packages
│   ├── handler/          # HTTP handlers
│   ├── repository/       # Database operations
│   ├── models/           # Data models
│   └── config/           # Configuration
├── pkg/                  # Shared packages
│   ├── pdf/              # PDF generation
│   ├── qrcode/           # QR code generation
│   ├── onesender/        # WhatsApp client
│   └── utils/            # Utilities
├── face_service/         # Python face recognition service
├── migrations/           # Database migrations
├── views/                # HTML templates
│   └── mobile/           # Mobile operator pages
└── public/               # Static files (audio, photos)
```

## API Endpoints

### Public
- `GET /` - Landing page
- `GET /scan` - RFID scan page
- `GET /scan-face` - Face recognition scan
- `GET /scan-sholat` - Prayer attendance scan

### Authentication
- `POST /api/login` - Admin login
- `POST /api/logout` - Logout

### Attendance (IoT)
- `GET /api/attendance/record?rfid=...` - Record RFID attendance
- `POST /api/attendance/verify-face` - Verify face for attendance
- `GET /api/attendance/today-stats` - Today's statistics
- `GET /api/attendance/recent-logs` - Recent attendance logs

### Sync
- `GET /api/sync` - Get schedules and announcements for device

### Admin (require auth)
- `GET /admin` - Dashboard
- `POST /admin/student/*` - Student CRUD
- `POST /admin/staff/*` - Staff CRUD
- `POST /admin/schedule/*` - Schedule CRUD
- `POST /admin/audio/*` - Audio file management
- `GET /admin/point-rules` - Point rules
- `POST /admin/points/transaction` - Add point transaction

## Setup

### Prerequisites
- Go 1.25+
- Python 3.8+ (untuk face_service)
- SQLite

### Build

```bash
# ARM64 (untuk server)
bash build_arm.sh

# Linux
bash build_linux.sh
```

### Deploy

```bash
bash deploy.sh
```

### Face Service Setup

```bash
cd face_service
pip install -r requirements.txt
python main.py
```

## Konfigurasi

Pengaturan tersedia di dashboard admin:
- **Jam Absensi**: arrival_start, arrival_end, departure_start, departure_end
- **WhatsApp**: OneSender API URL dan token
- **Template WA**: Custom message templates dengan variabel {name}, {time}, {status}
- **Jam Sholat**: dzuhur_start, dzuhur_end, ashar_start, ashar_end
- **Hari Libur**: Pengaturan hari kerja dan holiday

## Lisensi

MIT
