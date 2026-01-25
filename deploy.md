# Panduan Deployment SmartBell ke VPS (Lengkap)

## 1. Persiapan File
Di komputer lokal Anda, buka terminal di folder project dan jalankan build:

```powershell
$Env:GOOS = "linux"; $Env:GOARCH = "amd64"; go build -o bell_linux main.go
```

## 2. Upload ke VPS
Gunakan FileZilla/WinSCP untuk upload file berikut ke VPS (misal folder `/opt/smartbell`):
1. `bell_linux` (Binary aplikasi)
2. `setup.sh` (Wizard instalasi service)
3. `setup_nginx.sh` (Wizard domain & SSL)
4. Folder `views/`
5. Folder `public/`

## 3. Instalasi Aplikasi (Port 4000)
Login SSH ke VPS, lalu jalankan:

```bash
cd /opt/smartbell
bash setup.sh
```
*   Saat diminta Port, masukkan: **4000**
*   Tunggu sampai muncul "INSTALLASI BERHASIL".

## 4. Setup Domain & SSL (fo.bersekola.app)
Setelah aplikasi jalan di port 4000, jalankan script kedua dengan **sudo**:

```bash
sudo bash setup_nginx.sh
```
*   Masukkan Domain: `fo.bersekola.app`
*   Masukkan Port: `4000`
*   Pilih **y** saat ditanya install SSL.
*   Ikuti instruksi Certbot (masukkan email, pilih Agree, pilih Redirect/2).

Selesai! Aplikasi Anda sudah online di `https://fo.bersekola.app`.
