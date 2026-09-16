# Changelog: Integrasi Gowa WhatsApp Gateway

**Tanggal:** 2026-09-16  
**Versi:** v2.1.0

## Ringkasan

NIBA SuperApps v2 sekarang mendukung **Gowa (go-whatsapp-web-multidevice)** sebagai WhatsApp Gateway, dengan backward compatibility ke format OneSender lama.

## Perubahan

### 1. Package `pkg/onesender/client.go`

- ✅ Tambah field `Username` dan `DeviceID` ke struct `Client`
- ✅ Tambah struct `GowaTextPayload` dan `GowaImagePayload` untuk format Gowa
- ✅ Tambah struct `GowaResponse` untuk response Gowa
- ✅ Ubah `NewClient()` untuk terima optional `username` dan `deviceID`
- ✅ Tambah fungsi `normalizePhone()` untuk format nomor 62xxx
- ✅ Tambah fungsi `sendGowaText()` dan `sendGowaImage()` untuk Gowa API
- ✅ Update `SendTextMessage()` dan `SendImageMessage()` untuk deteksi mode Gowa/OneSender

### 2. Service `internal/service/wa.go`

- ✅ Update `Send()` untuk baca 4 setting: url, token, username, device_id
- ✅ Update `Test()` untuk baca 4 setting dan passing ke client

### 3. Handler `internal/handler/attendance.go`

- ✅ Tambah field `onesender_username` dan `onesender_device_id` ke form handler
- ✅ Update `UpdateAttendanceSettings()` untuk save 2 field baru

### 4. Database `internal/repository/db.go`

- ✅ Tambah insert default `onesender_username` dan `onesender_device_id` di `seedData()`
- ✅ Tambah insert default di `seedMySQLData()` (MySQL compatibility)

### 5. View `views/admin.html`

- ✅ Ubah label "OneSender" → "Gowa Gateway"
- ✅ Tambah input field "Username (Basic Auth)"
- ✅ Tambah input field "Device ID"
- ✅ Update placeholder URL: `http://127.0.0.1:8053`

### 6. Main `main.go`

- ✅ Tambah import `encoding/base64` untuk Basic Auth
- ✅ Tambah fungsi `normalizePhoneNumber()` untuk format 62xxx
- ✅ Update `SendOneSenderMessage()` untuk deteksi mode Gowa/OneSender
- ✅ Tambah fungsi `sendGowaMessage()` untuk kirim via Gowa API

### 7. Migration

- ✅ Buat file `migrations/000003_gowa_integration.up.sql` untuk tambah 2 field baru

## Cara Kerja

### Mode Detection

Client otomatis detect mode berdasarkan setting:

- **Gowa mode**: jika `onesender_username` dan `onesender_device_id` terisi
- **OneSender mode**: jika kedua field kosong (backward compatibility)

### Format Gowa

- **Endpoint text**: `POST /api/whatsapp/send`
- **Endpoint image**: `POST /api/whatsapp/send-image`
- **Auth**: Basic Auth (username:password → base64)
- **Header wajib**: `X-Device-Id`
- **Body JSON**: `{"phone":"628xxx","message":"..."}`
- **Nomor**: format `628xxx` tanpa + atau 0 di depan

### Format OneSender (lama)

- **Endpoint**: `POST {API_URL}`
- **Auth**: Bearer token di header
- **Body JSON**: `{"to":"...","recipient_type":"...","type":"text/image",...}`

## Testing

### 1. Update binary

```bash
# Build
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/smartbell_linux_arm64 -ldflags="-s -w"

# Upload ke VPS
scp dist/smartbell_linux_arm64 armbian@192.168.1.101:/opt/nibasuperappsv2/smartbell_linux_arm64.new

# Update via setup.sh
ssh armbian@192.168.1.101
cd /opt/nibasuperappsv2
sudo bash setup.sh
# Pilih menu "Update Aplikasi"
```

### 2. Jalankan migration SQL

```bash
sqlite3 /opt/nibasuperappsv2/database.db < migrations/000003_gowa_integration.up.sql
```

### 3. Set config via dashboard

Buka `http://192.168.1.101:8073/admin` → Settings:

- **API Base URL**: `http://127.0.0.1:8053`
- **Password (Basic Auth)**: `PutihAbu123!`
- **Username (Basic Auth)**: `admin`
- **Device ID**: `Pionir` atau `wa-broadcast`

### 4. Test kirim pesan

Buka `http://192.168.1.101:8073/admin` → Test WhatsApp → masukkan nomor test → kirim.

## Checklist Deployment

- [ ] Build binary baru
- [ ] Upload ke VPS
- [ ] Restart service
- [ ] Jalankan migration SQL
- [ ] Set config Gowa di dashboard
- [ ] Test kirim pesan
- [ ] Test tap RFID siswa
- [ ] Cek `whatsapp_logs` table untuk status

## Rollback

Jika perlu rollback ke OneSender:

1. Kosongkan field `onesender_username` dan `onesender_device_id` di dashboard
2. Isi kembali `onesender_api_url` dan `onesender_api_token` OneSender lama
3. System otomatis kembali ke mode OneSender

## Device Gowa yang Tersedia

| Device ID | Platform | JID | Status |
|---|---|---|---|
| Pionir | samsung | 6285158250766@s.whatsapp.net | Connected |
| wa-broadcast | android | 6285175449600@s.whatsapp.net | Connected |

## Catatan

- Backward compatibility: sistem tetap support OneSender jika field Gowa kosong
- Normalisasi nomor: semua nomor otomatis dinormalisasi ke format 62xxx
- Error handling: jika Gowa gagal, error dicatat ke `whatsapp_logs` table
- Multiple device: bisa switch device dengan ubah `Device ID` di dashboard
