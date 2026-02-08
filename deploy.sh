#!/bin/bash

# SmartBell Deployment Wizard
# All-in-one script for Installation, Update, and Domain Setup

echo "========================================="
echo "   SmartBell Deployment Wizard 🚀        "
echo "========================================="

# Check root
if [ "$EUID" -ne 0 ]; then
  echo "❌ Harap jalankan script ini dengan sudo (sudo bash deploy.sh)"
  exit 1
fi

PS3='Pilih menu (masukkan angka): '
options=("Install Baru (Fresh Install)" "Update Aplikasi (Setelah Upload)" "Setup Domain & SSL" "Keluar")
select opt in "${options[@]}"
do
    case $opt in
        "Install Baru (Fresh Install)")
            echo "--- Menjalankan Installer (setup.sh) ---"
            if [ -f "./setup.sh" ]; then
                bash ./setup.sh
            else
                echo "❌ File setup.sh tidak ditemukan!"
            fi
            break
            ;;
        "Update Aplikasi (Setelah Upload)")
            echo "--- Update Aplikasi ---"
            echo "[*] Menghentikan service..."
            systemctl stop smartbell
            
            echo "[*] Memperbarui izin eksekusi..."
            if [ -f "bell_linux" ]; then
                chmod +x bell_linux
            else
                echo "⚠️ File bell_linux tidak ditemukan! Pastikan sudah diupload."
            fi
            
            echo "[*] Menjalankan kembali service..."
            systemctl start smartbell
            
            echo "[*] Cek status..."
            sleep 2
            if systemctl is-active --quiet smartbell; then
                echo "✅ UPDATE BERHASIL! Aplikasi berjalan."
            else
                echo "❌ Gagal menjalankan aplikasi. Cek log dengan: sudo journalctl -u smartbell -n 20"
            fi
            break
            ;;
        "Setup Domain & SSL")
            echo "--- Setup Domain & SSL (setup_nginx.sh) ---"
            if [ -f "./setup_nginx.sh" ]; then
                bash ./setup_nginx.sh
            else
                echo "❌ File setup_nginx.sh tidak ditemukan!"
            fi
            break
            ;;
        "Keluar")
            echo "Bye! 👋"
            break
            ;;
        *) echo "Pilihan tidak valid $REPLY";;
    esac
done
