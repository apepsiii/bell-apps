#!/bin/bash

# SmartBell Setup Wizard
# Run this on your VPS to configure the app

echo "========================================="
echo "   SmartBell Installation Wizard 🔔      "
echo "========================================="

# 1. Check Files
echo "[*] Memeriksa file aplikasi..."
if [ ! -f "bell_linux" ]; then
    echo "❌ File 'bell_linux' tidak ditemukan!"
    echo "   Pastikan Anda sudah upload file binary hasil compile."
    exit 1
fi

if [ ! -d "views" ]; then
    echo "❌ Folder 'views' tidak ditemukan!"
    echo "   Pastikan Anda sudah upload folder views."
    exit 1
fi

# 2. Permission & Folders
echo "[*] Mengatur izin dan folder..."
chmod +x bell_linux
mkdir -p public/assets/audio
mkdir -p public/assets/photos
mkdir -p tmp

# 3. Configuration Input
echo ""
echo "--- Konfigurasi ---"
read -p "Masukkan PORT yang ingin digunakan (Default: 8080): " APP_PORT
APP_PORT=${APP_PORT:-8080}

echo "Aplikasi akan berjalan di Port: $APP_PORT"

# 4. Create Systemd Service
SERVICE_FILE="/etc/systemd/system/smartbell.service"
CURRENT_DIR=$(pwd)

echo "[*] Membuat Service Systemd di $SERVICE_FILE..."

# Need sudo for writing to /etc
cat <<EOF | sudo tee $SERVICE_FILE > /dev/null
[Unit]
Description=SmartBell Attendance Server
After=network.target

[Service]
User=root
WorkingDirectory=$CURRENT_DIR
ExecStart=$CURRENT_DIR/bell_linux
Restart=always
Environment=PORT=$APP_PORT

[Install]
WantedBy=multi-user.target
EOF

# 5. Start Service
echo "[*] Mengaktifkan Service..."
sudo systemctl daemon-reload
sudo systemctl enable smartbell
sudo systemctl restart smartbell

# 6. Check Status
echo "[*] Menunggu aplikasi start..."
sleep 2
STATUS=$(sudo systemctl is-active smartbell)

if [ "$STATUS" == "active" ]; then
    echo ""
    echo "✅ INSTALLASI BERHASIL!"
    echo "========================================="
    echo "Aplikasi berjalan di: http://IP_VPS_ANDA:$APP_PORT"
    echo "Untuk cek log: sudo journalctl -u smartbell -f"
    echo "========================================="
else
    echo ""
    echo "⚠️  Gagal menjalankan service."
    echo "Cek error dengan: sudo journalctl -u smartbell -n 20"
fi
