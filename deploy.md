# Panduan Deployment SmartBell ke VPS

Panduan ini menjelaskan langkah-langkah untuk mengonlinekan aplikasi SmartBell ke Virtual Private Server (VPS) berbasis Linux (Ubuntu/Debian).

## 1. Persiapan

Pastikan Anda memiliki:
- Akses SSH ke VPS (IP Address, Username, Password/Key).
- VPS dengan OS Ubuntu 20.04/22.04 atau Debian.
- Aplikasi FTP Client (FileZilla atau WinSCP) di komputer Anda.

## 2. Compile Aplikasi (Build)

Karena VPS menggunakan Linux, kita harus mengubah kode Go menjadi file program Linux (`binary`) dari komputer Windows Anda.

1. Buka Terminal / CMD / PowerShell di folder project `E:\smk\bell`.
2. Jalankan perintah berikut untuk compile:

   **PowerShell:**
   ```powershell
   $Env:GOOS = "linux"; $Env:GOARCH = "amd64"; go build -o bell_linux main.go
   ```

   **CMD (Command Prompt):**
   ```cmd
   set GOOS=linux
   set GOARCH=amd64
   go build -o bell_linux main.go
   ```

   **Git Bash:**
   ```bash
   GOOS=linux GOARCH=amd64 go build -o bell_linux main.go
   ```

3. Anda akan melihat file baru bernama `bell_linux` (tanpa ekstensi .exe) di folder project. Ini adalah aplikasi utama untuk server.

## 3. Upload File ke VPS

1. Buka FileZilla / WinSCP.
2. Login ke VPS Anda (Protocol: SFTP).
3. Buat folder baru di VPS, misalnya `/opt/smartbell`.
4. Upload file dan folder berikut ke dalam `/opt/smartbell`:
   - `bell_linux` (File hasil compile tadi)
   - `views/` (Folder template HTML - **Wajib**)
   - `public/` (Folder aset CSS/JS/Foto - **Wajib**)

   *> Catatan: Jangan upload file `.db` jika ingin mulai database dari nol. Jika ingin migrasi data yang sudah ada di lokal, upload juga `database.db`.*

Struktur di VPS harusnya terlihat seperti ini:
```
/opt/smartbell/
├── bell_linux
├── database.db (Opsional)
├── public/
└── views/
```

## 4. Menjalankan Aplikasi

1. Login SSH ke VPS (gunakan PuTTY atau Terminal).
2. Masuk ke folder aplikasi dan beri izin eksekusi pada file binary:
   ```bash
   cd /opt/smartbell
   chmod +x bell_linux
   ```
3. Coba jalankan aplikasi (Test Run):
   ```bash
   ./bell_linux
   ```
   Jika muncul log server started (misal `:8080`), berarti berhasil. Tekan `Ctrl + C` untuk berhenti.

## 5. Membuat Service (Agar Jalan Otomatis)

Agar aplikasi tetap jalan walau SSH ditutup dan otomatis hidup kembali saat VPS restart, kita gunakan `systemd`.

1. Buat file service baru:
   ```bash
   sudo nano /etc/systemd/system/smartbell.service
   ```

2. Paste konfigurasi berikut ke dalamnya:
   ```ini
   [Unit]
   Description=SmartBell Attendance Server
   After=network.target

   [Service]
   # User root atau user lain yg memiliki akses ke folder
   User=root
   # Folder tempat aplikasi berada (PENTING)
   WorkingDirectory=/opt/smartbell
   # Perintah menjalankan aplikasi
   ExecStart=/opt/smartbell/bell_linux
   # Restart otomatis jika crash
   Restart=always

   [Install]
   WantedBy=multi-user.target
   ```

3. Simpan file (Tekan `Ctrl+O`, `Enter`, lalu `Ctrl+X`).

4. Aktifkan dan jalankan service:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable smartbell
   sudo systemctl start smartbell
   ```

5. Cek status service untuk memastikan berjalan:
   ```bash
   sudo systemctl status smartbell
   ```
   Pastikan statusnya berwarna hijau **active (running)**.

Akses aplikasi di browser: `http://IP_VPS_ANDA:8080`

## 6. (Opsional) Menggunakan Domain & Nginx

Jika ingin menggunakan domain (misal `sekolah.sch.id`) dan menghilangkan port :8080 di URL, gunakan Nginx sebagai Reverse Proxy.

1. Install Nginx:
   ```bash
   sudo apt update
   sudo apt install nginx
   ```

2. Buat konfigurasi server block:
   ```bash
   sudo nano /etc/nginx/sites-available/smartbell
   ```

3. Isi dengan konfigurasi berikut:
   ```nginx
   server {
       listen 80;
       server_name domain-anda.com; # Ganti dengan domain atau IP VPS Anda

       location / {
           proxy_pass http://localhost:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           
           # Support WebSocket (penting untuk beberapa fitur realtime jika ada)
           proxy_http_version 1.1;
           proxy_set_header Upgrade $http_upgrade;
           proxy_set_header Connection "upgrade";
       }
   }
   ```

4. Aktifkan konfigurasi dan Restart Nginx:
   ```bash
   sudo ln -s /etc/nginx/sites-available/smartbell /etc/nginx/sites-enabled/
   sudo nginx -t
   sudo systemctl restart nginx
   ```

Sekarang aplikasi bisa diakses langsung di `http://domain-anda.com`.
