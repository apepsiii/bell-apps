#!/usr/bin/env bash
# ==============================================================================
# Script Build & Package Deployment VPS (Linux amd64)
# SMK NIBA Super Apps (bell-apps)
# ==============================================================================

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}======================================================${NC}"
echo -e "${CYAN}   SMK NIBA Super Apps - Build Deployment VPS Linux   ${NC}"
echo -e "${CYAN}======================================================${NC}"

DIST_DIR="dist"
APP_NAME="bell-apps"
PACKAGE_NAME="${APP_NAME}-vps-linux-amd64.tar.gz"

echo -e "\n${BLUE}[1/6] Membersihkan direktori build sebelumnya...${NC}"
rm -rf "${DIST_DIR}" "${PACKAGE_NAME}"
mkdir -p "${DIST_DIR}"
mkdir -p "${DIST_DIR}/logs"
mkdir -p "${DIST_DIR}/public/assets/audio"
mkdir -p "${DIST_DIR}/public/assets/photos"
mkdir -p "${DIST_DIR}/public/assets/signage"

echo -e "\n${BLUE}[2/6] Melakukan kompilasi Go binary untuk Linux (amd64)...${NC}"
echo "    - Target OS   : Linux"
echo "    - Target Arch : amd64"
echo "    - CGO         : Disabled (Pure Go SQLite)"

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

go build -ldflags="-s -w" -o "${DIST_DIR}/${APP_NAME}" .
chmod +x "${DIST_DIR}/${APP_NAME}" 2>/dev/null || true
echo -e "${GREEN}✓ Kompilasi sukses: ${DIST_DIR}/${APP_NAME}${NC}"

echo -e "\n${BLUE}[3/6] Menyalin aset dan konfigurasi...${NC}"
if [ -f "config.yaml" ]; then
    cp config.yaml "${DIST_DIR}/config.yaml"
    echo -e "    ✓ config.yaml disalin"
fi

if [ -d "public/assets/audio" ]; then
    cp -r public/assets/audio/* "${DIST_DIR}/public/assets/audio/" 2>/dev/null || true
    echo -e "    ✓ Audio assets disalin"
fi

touch "${DIST_DIR}/logs/.gitkeep"
touch "${DIST_DIR}/public/assets/photos/.gitkeep"
touch "${DIST_DIR}/public/assets/signage/.gitkeep"

echo -e "\n${BLUE}[4/6] Membuat template Systemd Service...${NC}"
cat << 'EOF' > "${DIST_DIR}/${APP_NAME}.service"
[Unit]
Description=SMK NIBA Super Apps Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/bell-apps
ExecStart=/var/www/bell-apps/bell-apps
Restart=always
RestartSec=5
LimitNOFILE=65535
Environment=APP_VERSION=2.1.0
Environment=APP_NAME=NIBA SuperApps

[Install]
WantedBy=multi-user.target
EOF
echo -e "    ✓ File systemd dibuat: ${DIST_DIR}/${APP_NAME}.service"

echo -e "\n${BLUE}[5/6] Membuat script helper installer di VPS...${NC}"
cat << 'EOF' > "${DIST_DIR}/install.sh"
#!/usr/bin/env bash
# Script instalasi otomatis di VPS
set -e

INSTALL_DIR="/var/www/bell-apps"
SERVICE_NAME="bell-apps"

echo "=== Memulai Instalasi ${SERVICE_NAME} di VPS ==="

# 1. Buat direktori tujuan
mkdir -p "${INSTALL_DIR}"
mkdir -p "${INSTALL_DIR}/logs"
mkdir -p "${INSTALL_DIR}/public/assets/audio"
mkdir -p "${INSTALL_DIR}/public/assets/photos"
mkdir -p "${INSTALL_DIR}/public/assets/signage"

# 2. Salin binary & file pendukung jika dijalankan dari luar INSTALL_DIR
if [ "$(pwd)" != "${INSTALL_DIR}" ]; then
    echo "Menyalin file ke ${INSTALL_DIR}..."
    cp -f bell-apps "${INSTALL_DIR}/"
    [ -f config.yaml ] && cp -n config.yaml "${INSTALL_DIR}/"
    cp -rn public "${INSTALL_DIR}/" 2>/dev/null || true
fi

# 3. Beri permission eksekusi
chmod +x "${INSTALL_DIR}/bell-apps"

# 4. Pasang Systemd Service
if [ -f "${SERVICE_NAME}.service" ]; then
    echo "Memasang systemd service..."
    cp "${SERVICE_NAME}.service" "/etc/systemd/system/${SERVICE_NAME}.service"
    systemctl daemon-reload
    systemctl enable "${SERVICE_NAME}"
    systemctl restart "${SERVICE_NAME}"
    echo "Service ${SERVICE_NAME} berhasil diaktifkan dan dijalankan!"
    echo "Cek status dengan: systemctl status ${SERVICE_NAME}"
else
    echo "File service tidak ditemukan, jalankan manual: cd ${INSTALL_DIR} && ./bell-apps"
fi

echo "=== Instalasi Selesai! ==="
EOF
chmod +x "${DIST_DIR}/install.sh" 2>/dev/null || true
echo -e "    ✓ Script installer VPS dibuat: ${DIST_DIR}/install.sh"

echo -e "\n${BLUE}[6/6] Mengompresi paket distribusi menjadi arsip tar.gz...${NC}"
# Buat arsip di luar dist lalu pindahkan ke dalam dist
tar -czf "${PACKAGE_NAME}" -C "${DIST_DIR}" .
mv "${PACKAGE_NAME}" "${DIST_DIR}/${PACKAGE_NAME}"

echo -e "${GREEN}✓ Arsip berhasil dibuat: ${DIST_DIR}/${PACKAGE_NAME}${NC}"

echo -e "\n${GREEN}======================================================${NC}"
echo -e "${GREEN}   BUILD SELESAI DENGAN SUKSES!                       ${NC}"
echo -e "${GREEN}======================================================${NC}"
echo -e "Folder hasil: ${YELLOW}${DIST_DIR}/${NC}"
echo -e "File arsip  : ${YELLOW}${DIST_DIR}/${PACKAGE_NAME}${NC}"
echo -e ""
echo -e "Cara upload & jalankan di VPS:"
echo -e "  1. Upload ${YELLOW}${DIST_DIR}/${PACKAGE_NAME}${NC} ke VPS (via scp/sftp/cPanel):"
echo -e "     ${CYAN}scp ${DIST_DIR}/${PACKAGE_NAME} root@IP_VPS:/tmp/${NC}"
echo -e "  2. Di terminal VPS, ekstrak dan jalankan installer:"
echo -e "     ${CYAN}mkdir -p /var/www/bell-apps && tar -xzf /tmp/${PACKAGE_NAME} -C /var/www/bell-apps${NC}"
echo -e "     ${CYAN}cd /var/www/bell-apps && bash install.sh${NC}"
echo -e ""
