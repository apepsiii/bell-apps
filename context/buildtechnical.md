# Build Technical Documentation - SmartBell ARM Build

## Overview

`build_arm.sh` adalah script Bash yang digunakan untuk mengkompilasi binary SmartBell specifically untuk arsitektur ARM64, yang digunakan pada perangkat Armbian seperti Orange Pi dan Rock Pi.

---

## Prerequisites

- **Go 1.25+** terinstall
- **Bash shell** (di Linux/macOS atau Git Bash di Windows)
- Akses ke source code `main.go` yang mengandung variabel `AppVersion`

---

## Environment Variables Build

Script ini menggunakan **cross-compilation** dengan menset dua environment variables:

| Variable | Value | Purpose |
|----------|-------|---------|
| `GOOS` | `linux` | Target OS adalah Linux |
| `GOARCH` | `arm64` | Target arsitektur ARM 64-bit |

### Kenapa Cross-Compilation?

Dengan cross-compilation, kita dapat mengkompilasi binary ARM64 langsung dari mesin Windows atau Linux x86/64 tanpa perlu emulator atau perangkat ARM fisik. Go memiliki built-in support untuk cross-compilation ke berbagai arsitektur.

---

## Proses Build

### 1. Ekstraksi Version

```bash
VERSION=$(grep 'AppVersion.*=' main.go | grep -oP 'v[0-9]+\.[0-9]+\.[0-9]+' | tr '.' '_')
```

Penjelasan step-by-step:

1. **`grep 'AppVersion.*=' main.go`** - Mencari baris yang mengandung `AppVersion` dan tanda `=`
2. **`grep -oP 'v[0-9]+\.[0-9]+\.[0-9]+'`** - Ekstrak pola version (contoh: `v1.2.1`) menggunakan regex Perl-compatible
3. **`tr '.' '_'`** - Mengganti titik dengan underscore, menghasilkan `v1_2_1`

Contoh hasil: `v1.2.1` -> `v1_2_1`

### 2. Generate Tanggal

```bash
DATE=$(date +"%d%m%y")
```

Format: `ddmmyy` (2 digit tanggal, bulan, tahun)

Contoh hasil: `100226` (10 February 2026)

### 3. Nama Output File

```bash
OUT_FILE="smartbell_${VERSION}_${DATE}_arm64"
```

Contoh hasil: `smartbell_v1_2_1_100226_arm64`

### 4. Perintah Build

```bash
GOOS=linux GOARCH=arm64 go build -o "$OUT_FILE" .
```

- `-o "$OUT_FILE"` - Menentukan nama output file
- `.` - Build dari current directory (main.go sebagai entry point)

---

## Perbandingan: ARM vs Linux x86/64

| Aspek | `build_arm.sh` | `build_linux.sh` |
|-------|---------------|------------------|
| `GOARCH` | `arm64` | `amd64` |
| Output suffix | `_arm64` | (tanpa suffix) |
| Target devices | Orange Pi, Rock Pi, Armbian | VPS (DigitalOcean, AWS, dll) |
| Use case | Edge device/single board | Server/cloud |

---

## Contoh Penggunaan

### Build ARM Binary (dari Windows/Linux/macOS):

```bash
# Clone repo
git clone https://github.com/username/bell.git
cd bell

# Jalankan script build
bash build_arm.sh
```

### Output:

```
========================================
   SmartBell Build for Armbian (ARM) 🐧
========================================

[*] Compiling binary for ARM64...
[*] Version: v1_2_1
[*] Output: smartbell_v1_2_1_100226_arm64

✅ BUILD SUCCESS!
File 'smartbell_v1_2_1_100226_arm64' siap diupload ke VPS Armbian.
```

---

## Deployment ke Armbian

Setelah binary berhasil di-build:

```bash
# Upload ke VPS ARM (gunakan SCP atau FileZilla)
scp smartbell_v1_2_1_100226_arm64 root@<ip-vps>:/opt/smartbell/

# SSH ke VPS
ssh root@<ip-vps>

# Buat executable dan replace
chmod +x smartbell_v1_2_1_100226_arm64
cp smartbell_v1_2_1_100226_arm64 /opt/smartbell/bell_linux

# Restart service
systemctl restart smartbell
```

---

## Troubleshooting

### Error: `command not found: go`
Go belum terinstall atau tidak ada di PATH. Install Go terlebih dahulu:
```bash
# Linux
sudo apt install golang

# Verifikasi
go version
```

### Error: `build failed`
Cek apakah `main.go` ada di current directory dan tidak ada syntax error:
```bash
go build -o test_build .
```

### Build berhasil tapi binary tidak jalan
Pastikan binary sudah executable:
```bash
chmod +x smartbell_v1_2_1_100226_arm64
```

---

## Integrasi dengan deploy.sh

`build_arm.sh` menghasilkan file yang kemudian digunakan oleh `deploy.sh`. Di dalam `deploy.sh`, script会自动 mendeteksi file ARM terbaru:

```bash
BINARY_FILE=$(ls -t smartbell_v*_arm64 2>/dev/null | head -1)
```

Dan mengcopy ke `/opt/smartbell/bell_linux` untuk dijalankan sebagai service.

---

## File Pendukung Deployment

| File | Fungsi |
|------|--------|
| `build_arm.sh` | Compile binary ARM64 |
| `setup.sh` | Install service systemd |
| `setup_nginx.sh` | Setup reverse proxy + SSL |
| `deploy.sh` | Wizard deployment menu |

---

## Catatan Teknis Tambahan

- **No CGO**: Build dilakukan dengan CGO disabled (default) karena SQLite driver yang digunakan (`modernc.org/sqlite`) adalah pure-Go
- **Static Binary**: Binary yang di-build adalah static binary yang tidak memerlukan shared libraries
- **Portability**: Binary ARM64 dapat dijalankan di berbagai device Armbian tanpa instalasi dependencies tambahan
