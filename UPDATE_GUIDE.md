# Update Guide untuk SmartBell v1.2.0

## Cara Update Aplikasi di Armbian

### 1. Build Binary di Windows/Mac

```bash
# Jalankan build script untuk ARM64
bash build_arm.sh
```

Ini akan menghasilkan file: `smartbell_v1_2_0_DDMMYY_arm64`

### 2. Upload ke Server Armbian

```bash
# Upload file binary ke server
scp smartbell_v1_2_0_*_arm64 user@server-ip:/home/user/
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

- ✅ Menghentikan service smartbell
- ✅ Mencari file binary terbaru (`smartbell_v*_arm64`)
- ✅ Copy ke `/opt/smartbell/bell_linux`
- ✅ Set permission executable
- ✅ Restart service
- ✅ Verifikasi status

## Troubleshooting

### File tidak ditemukan

Pastikan file binary sudah diupload ke direktori yang sama dengan `deploy.sh`

### Service gagal start

Cek log dengan:

```bash
sudo journalctl -u smartbell -n 20
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
sudo systemctl status smartbell
```

3. Cek log real-time:

```bash
sudo journalctl -u smartbell -f
```
