#!/bin/bash

# ============================================================
#  SMK NIBA Super Apps - Setup Wizard
#
#  All-in-one installer for first install, update, and domain/SSL.
#  Run on the VPS (as root) in the folder containing the binary:
#
#    cd /opt/nibasuperapps
#    sudo bash setup.sh
#
#  The script:
#    - Installs: creates systemd service, sets permissions, starts app
#    - Updates: backup old binary, swap new, restart, health-check
#    - Domain/SSL: nginx reverse proxy + certbot (placeholder)
# ============================================================

# --- Detect this app's identity from the script location ---
APP_DIR="$(cd "$(dirname "$0")" && pwd)"
APP_NAME="$(basename "$APP_DIR")"   # e.g. nibasuperapps
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"
SERVICE_NAME="$APP_NAME"

echo "========================================="
echo "   SMK NIBA Super Apps - Setup Wizard 🚀"
echo "========================================="
echo ""
echo "  App folder  : $APP_DIR"
echo "  Service name: $SERVICE_NAME"
echo ""

# Check root
if [ "$EUID" -ne 0 ]; then
  echo "❌ Harap jalankan dengan sudo:  sudo bash setup.sh"
  exit 1
fi

# --- Find the binary: prefer a bare `smartbell`/`nibasuperapps`,
# --- otherwise the newest smartbell_linux_* file. ---
detect_binary() {
    local preferred
    for c in smartbell nibasuperapps app; do
        if [ -f "$APP_DIR/$c" ] && [ -x "$APP_DIR/$c" ]; then
            echo "$c"
            return
        fi
    done
    # newest smartbell_linux_<arch>
    local arch
    arch=$(uname -m 2>/dev/null)
    case "$arch" in
        x86_64)        arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *)             arch="" ;;
    esac
    if [ -n "$arch" ]; then
        local found
        found=$(ls -t "$APP_DIR"/smartbell_linux_${arch} 2>/dev/null | head -1)
        if [ -n "$found" ]; then echo "$(basename "$found")"; return; fi
    fi
    found=$(ls -t "$APP_DIR"/smartbell_linux_* 2>/dev/null | head -1)
    if [ -n "$found" ]; then echo "$(basename "$found")"; return; fi
    echo ""
}

CURRENT_BIN=$(detect_binary)
if [ -z "$CURRENT_BIN" ]; then
    echo "❌ Binary tidak ditemukan di $APP_DIR"
    echo "   Upload binary (smartbell_linux_amd64 / arm64) ke folder ini."
    exit 1
fi
echo "[*] Binary terdeteksi: $CURRENT_BIN"
echo ""

# --- Menu ---
PS3='Pilih menu (masukkan angka): '
options=("Install Baru (Fresh Install)" "Update Aplikasi (Setelah Upload Binary Baru)" "Cek Status & Log" "Keluar")
select opt in "${options[@]}"
do
    case $opt in
    "Install Baru (Fresh Install)")
        echo ""
        echo "--- Install Baru ---"

        # Permissions & folders
        echo "[*] Mengatur izin dan folder..."
        chmod +x "$APP_DIR/$CURRENT_BIN"
        mkdir -p "$APP_DIR/public/assets/audio" "$APP_DIR/public/assets/photos" "$APP_DIR/logs" "$APP_DIR/tmp"

        # Config input
        echo ""
        read -p "Masukkan PORT yang ingin digunakan (Default: 8080): " APP_PORT
        APP_PORT=${APP_PORT:-8080}
        echo "Aplikasi akan berjalan di Port: $APP_PORT"

        # Write port into config.yaml (the app reads port from there)
        if [ -f "$APP_DIR/config.yaml" ]; then
            echo "[*] Update port di config.yaml -> $APP_PORT"
            sed -i "s/^    port: .*/    port: \"$APP_PORT\"/" "$APP_DIR/config.yaml" 2>/dev/null || true
        fi

        # Create systemd service
        echo "[*] Membuat Service Systemd: $SERVICE_FILE"
        cat <<EOF | tee "$SERVICE_FILE" > /dev/null
[Unit]
Description=SMK NIBA Super Apps Attendance Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/$CURRENT_BIN
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

        # Start service
        echo "[*] Mengaktifkan & memulai service..."
        systemctl daemon-reload
        systemctl enable "$SERVICE_NAME"
        systemctl restart "$SERVICE_NAME"

        # Check status
        echo "[*] Menunggu aplikasi start..."
        sleep 3
        STATUS=$(systemctl is-active "$SERVICE_NAME")
        if [ "$STATUS" == "active" ]; then
            echo ""
            echo "✅ INSTALLASI BERHASIL!"
            echo "========================================="
            echo "  Aplikasi : http://IP_VPS_ANDA:$APP_PORT"
            echo "  Login    : admin / admin123"
            echo "  Service  : systemctl status $SERVICE_NAME"
            echo "  Log      : journalctl -u $SERVICE_NAME -f"
            echo "             tail -f $APP_DIR/logs/app.log"
            echo "========================================="
        else
            echo ""
            echo "⚠️  Gagal menjalankan service."
            echo "Cek error: journalctl -u $SERVICE_NAME -n 30"
        fi
        break
        ;;

    "Update Aplikasi (Setelah Upload Binary Baru)")
        echo ""
        echo "--- Update Aplikasi ---"

        # Find new binary (any smartbell_linux_* newer than current, or a .new file)
        NEW_BIN=""
        for c in "$APP_DIR/smartbell.new" "$APP_DIR/$CURRENT_BIN.new"; do
            if [ -f "$c" ]; then NEW_BIN="$c"; break; fi
        done
        if [ -z "$NEW_BIN" ]; then
            # newest smartbell_linux_* that is different from current
            arch=$(uname -m 2>/dev/null)
            case "$arch" in
                x86_64)        arch="amd64" ;;
                aarch64|arm64) arch="arm64" ;;
                *)             arch="" ;;
            esac
            if [ -n "$arch" ]; then
                NEW_BIN=$(ls -t "$APP_DIR"/smartbell_linux_${arch} 2>/dev/null | head -1)
            fi
            if [ -z "$NEW_BIN" ]; then
                NEW_BIN=$(ls -t "$APP_DIR"/smartbell_linux_* 2>/dev/null | head -1)
            fi
        fi

        if [ -z "$NEW_BIN" ] || [ ! -f "$NEW_BIN" ]; then
            echo "❌ Binary baru tidak ditemukan."
            echo "   Upload smartbell_linux_<arch> (atau rename ke smartbell.new) di $APP_DIR"
            break
        fi

        echo "[*] Binary baru: $NEW_BIN"
        chmod +x "$NEW_BIN"

        # Backup old, stop, swap, start
        BACKUP="$APP_DIR/$CURRENT_BIN.bak.$(date +%Y%m%d%H%M%S)"
        echo "[*] Stop service + backup binary lama -> $(basename "$BACKUP")"
        systemctl stop "$SERVICE_NAME"
        cp "$APP_DIR/$CURRENT_BIN" "$BACKUP"

        # If the new binary has a different name, we need to either rename it
        # to match the service ExecStart, or update the service file.
        NEW_NAME="$(basename "$NEW_BIN")"
        if [ "$NEW_NAME" != "$CURRENT_BIN" ]; then
            echo "[*] Rename binary baru -> $CURRENT_BIN (sesuai service)"
            cp "$NEW_BIN" "$APP_DIR/$CURRENT_BIN"
            chmod +x "$APP_DIR/$CURRENT_BIN"
        else
            cp "$NEW_BIN" "$APP_DIR/$CURRENT_BIN"
            chmod +x "$APP_DIR/$CURRENT_BIN"
        fi

        echo "[*] Start service..."
        systemctl start "$SERVICE_NAME"

        echo "[*] Health check..."
        sleep 3
        if systemctl is-active --quiet "$SERVICE_NAME"; then
            echo "✅ UPDATE BERHASIL! Aplikasi jalan dengan binary baru."
            echo "   Backup binary lama: $(basename "$BACKUP")"
            echo "   Hapus backup jika sudah yakin: rm $(basename "$BACKUP")"
            echo "   Cek log: journalctl -u $SERVICE_NAME -f"
        else
            echo "❌ Gagal start dengan binary baru. Rollback..."
            systemctl stop "$SERVICE_NAME"
            cp "$BACKUP" "$APP_DIR/$CURRENT_BIN"
            chmod +x "$APP_DIR/$CURRENT_BIN"
            systemctl start "$SERVICE_NAME"
            sleep 2
            if systemctl is-active --quiet "$SERVICE_NAME"; then
                echo "⚠️  Binary lama dipulihkan. Aplikasi kembali jalan."
                echo "    Binary baru mungkin tidak kompatibel. Cek log:"
                echo "    journalctl -u $SERVICE_NAME -n 30"
            else
                echo "❌ Rollback juga gagal. Cek manual:"
                echo "    journalctl -u $SERVICE_NAME -n 30"
            fi
        fi
        break
        ;;

    "Cek Status & Log")
        echo ""
        echo "--- Status Service: $SERVICE_NAME ---"
        systemctl status "$SERVICE_NAME" --no-pager -l 2>/dev/null | head -20
        echo ""
        echo "--- 20 log terakhir ---"
        journalctl -u "$SERVICE_NAME" -n 20 --no-pager 2>/dev/null
        echo ""
        echo "--- app.log (10 baris) ---"
        tail -n 10 "$APP_DIR/logs/app.log" 2>/dev/null || echo "(file log belum ada)"
        break
        ;;

    "Keluar")
        echo "Bye! 👋"
        break
        ;;

    *) echo "Pilihan tidak valid.";;
    esac
done
