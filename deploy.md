# Panduan Deployment SmartBell ke VPS (Metode Wizard)

## 1. Compile Aplikasi (Di Windows)
Buka terminal di folder project dan jalankan:
```powershell
$Env:GOOS = "linux"; $Env:GOARCH = "amd64"; go build -o bell_linux main.go
```

## 2. Upload File
Upload file-file berikut ke VPS (misal ke folder `/opt/smartbell`):
1. `bell_linux` (Binary)
2. `setup.sh` (Script Wizard)
3. Folder `views/`
4. Folder `public/`

## 3. Jalankan Wizard di VPS
Login SSH ke VPS, masuk ke folder, dan jalankan wizard:

```bash
cd /opt/smartbell
bash setup.sh
```

Wizard akan menanyakan Port yang ingin digunakan dan otomatis mengatur semuanya (Service, Permission, Folder).

Selesai! Aplikasi langsung online.
