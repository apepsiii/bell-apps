#!/bin/bash

# SmartBell Domain & SSL Setup Wizard
# Gunakan script ini untuk menghubungkan Domain ke Aplikasi dan pasang SSL

echo "========================================="
echo "   SmartBell Domain & SSL Wizard 🌐      "
echo "========================================="

# 1. Cek Root
if [ "$EUID" -ne 0 ]; then
  echo "❌ Harap jalankan script ini dengan sudo (sudo bash setup_nginx.sh)"
  exit 1
fi

# 2. Input Konfigurasi
read -p "Masukkan DOMAIN Anda (contoh: fo.bersekola.app): " DOMAIN
if [ -z "$DOMAIN" ]; then
    echo "❌ Domain tidak boleh kosong!"
    exit 1
fi

read -p "Masukkan PORT Aplikasi yang sedang berjalan (contoh: 4000): " APP_PORT
APP_PORT=${APP_PORT:-4000}

# 3. Install Nginx & Certbot
echo ""
echo "[*] Menginstall Nginx dan Certbot..."
apt update
apt install -y nginx certbot python3-certbot-nginx

# 4. Buat Konfigurasi Nginx
CONFIG_FILE="/etc/nginx/sites-available/$DOMAIN"

echo "[*] Membuat konfigurasi Nginx di $CONFIG_FILE..."

cat <<EOF > $CONFIG_FILE
server {
    server_name $DOMAIN;

    location / {
        proxy_pass http://localhost:$APP_PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        # WebSocket Support (Penting untuk realtime jika ada)
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF

# 5. Aktifkan Config
if [ ! -f "/etc/nginx/sites-enabled/$DOMAIN" ]; then
    ln -s $CONFIG_FILE /etc/nginx/sites-enabled/
fi

# Cek Config
nginx -t
if [ $? -ne 0 ]; then
    echo "❌ Konfigurasi Nginx Error. Cek pesan di atas."
    exit 1
fi

echo "[*] Restarting Nginx..."
systemctl restart nginx

# 6. Pasang SSL (Certbot)
echo ""
echo "========================================="
read -p "Apakah Anda ingin install SSL (HTTPS) sekarang? (y/n): " INSTALL_SSL
if [ "$INSTALL_SSL" == "y" ]; then
    echo "[*] Menjalankan Certbot..."
    echo "    (Ikuti instruksi di layar: Masukkan Email & Pilih Redirect)"
    certbot --nginx -d $DOMAIN
fi

echo ""
echo "✅ SETUP SELESAI!"
echo "Akses aplikasi di: https://$DOMAIN"
