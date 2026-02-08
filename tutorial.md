# Panduan Deployment SmartBell (Single Binary) 🚀

Dokumen ini menjelaskan cara deployment menggunakan **satu file aplikasi saja**.
Semua script setup dan tampilan HTML sudah menyatu di dalam file `bell_linux`.

---

## 1️⃣ Tahap 1: Build di Local (Windows)

1.  Buka **File Explorer** di folder project.
2.  Double-click **`build_linux.bat`**.
3.  Tunggu hingga muncul file **`bell_linux`**.

---

## 2️⃣ Tahap 2: Upload File (Cukup 1 File!)

Anda hanya perlu mengupload **Satu File**: `bell_linux`.

### Cara Upload (Contoh via transfer.sh)

1.  Buka terminal/Git Bash di folder project.
2.  Upload file:
    ```bash
    curl --upload-file ./bell_linux https://transfer.sh/bell_linux
    ```
3.  **Simpan Link Download** yang muncul!

---

## 3️⃣ Tahap 3: Download & Deploy di VPS

Login ke VPS Anda, lalu jalankan perintah berikut:

### 1. Download File

```bash
# Hapus file lama (opsional tapi disarankan agar bersih)
rm -f bell_linux

# Download file baru (Ganti URL dengan link Anda)
curl -L -o bell_linux "https://transfer.sh/LinkAnda/bell_linux"

# Berikan izin eksekusi
chmod +x bell_linux
```

### 2. Jalankan Wizard

Jalankan aplikasi dalam mode wizard untuk setup atau update.

```bash
sudo ./bell_linux wizard
```

Menu yang tersedia:

- **1) Install Baru**: Untuk pemasangan awal (membuat service systemd, folder, dll).
- **2) Update Aplikasi**: Pilih ini jika Anda baru saja menimpa file untuk update.
- **3) Setup Domain & SSL**: Untuk menghubungkan domain dan HTTPS.

---

## ✅ Selesai!

Aplikasi berjalan otomatis. Cek log dengan:

```bash
sudo journalctl -u smartbell -f
```
