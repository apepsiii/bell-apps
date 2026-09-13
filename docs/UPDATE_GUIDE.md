# Update Guide untuk SMK NIBA Super Apps v1.2.0

## Cara Update Aplikasi di Armbian

### 1. Build Binary di Windows/Mac

```bash
# Jalankan build script untuk ARM64
bash build_arm.sh
```

Ini akan menghasilkan file: `SMK NIBA Super Apps_v1_2_0_DDMMYY_arm64`

### 2. Upload ke Server Armbian

```bash
# Upload file binary ke server
scp SMK NIBA Super Apps_v1_2_0_*_arm64 user@server-ip:/home/user/
```

### 3. Jalankan Deploy Wizard

```bash
# SSH ke server
ssh user@server-ip

# Masuk ke direktori aplikasi
cd /home/user/

# Jalankan deploy wizard dengan sudo
sudo bash deploy.sh
```

### 4. Pilih Menu Update

```
Pilih menu (masukkan angka): 2
```

Script akan otomatis:

- ✅ Menghentikan service SMK NIBA Super Apps
- ✅ Mencari file binary terbaru (`SMK NIBA Super Apps_v*_arm64`)
- ✅ Copy ke `/opt/SMK NIBA Super Apps/bell_linux`
- ✅ Set permission executable
- ✅ Restart service
- ✅ Verifikasi status

## Troubleshooting

### File tidak ditemukan

Pastikan file binary sudah diupload ke direktori yang sama dengan `deploy.sh`

### Service gagal start

Cek log dengan:

```bash
sudo journalctl -u SMK NIBA Super Apps -n 20
```

### Permission denied

Pastikan menjalankan dengan sudo:

```bash
sudo bash deploy.sh
```

## Verifikasi Update Berhasil

1. Cek versi aplikasi di browser: http://server-ip:8080
2. Cek status service:

```bash
sudo systemctl status SMK NIBA Super Apps
```

3. Cek log real-time:

```bash
sudo journalctl -u SMK NIBA Super Apps -f
```
