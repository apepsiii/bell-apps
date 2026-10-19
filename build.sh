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
# Format penamaan dengan penanda waktu: TAHUNBULANTANGGAL_JAMMENIT (cth: bell-apps-vps-20261017_2135.zip)
TIMESTAMP=$(date +"%Y%m%d_%H%M")
ZIP_NAME="${APP_NAME}-vps-${TIMESTAMP}.zip"

echo -e "\n${BLUE}[1/6] Membersihkan direktori build sebelumnya...${NC}"
rm -rf "${DIST_DIR}" "${APP_NAME}-vps-"*.zip
mkdir -p "${DIST_DIR}"
mkdir -p "${DIST_DIR}/logs"
mkdir -p "${DIST_DIR}/public/assets/audio"
mkdir -p "${DIST_DIR}/public/assets/photos"
mkdir -p "${DIST_DIR}/public/assets/signage"

TARGET_ARCH="${1:-${GOARCH:-amd64}}"
VERSION=$(git describe --tags --always 2>/dev/null || grep 'Version.*=' version.go | grep -oP '"[0-9\.]+"' | tr -d '"' || echo "2.1.0")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo -e "\n${BLUE}[2/6] Melakukan kompilasi Go binary untuk Linux (${TARGET_ARCH})...${NC}"
echo "    - Target OS   : Linux"
echo "    - Target Arch : ${TARGET_ARCH}"
echo "    - CGO         : Disabled (Pure Go SQLite)"
echo "    - Version     : ${VERSION}"
echo "    - Git Commit  : ${GIT_COMMIT}"

export CGO_ENABLED=0
export GOOS=linux
export GOARCH="${TARGET_ARCH}"

go build -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildDate=${TIMESTAMP} -X main.GitCommit=${GIT_COMMIT}" -o "${DIST_DIR}/${APP_NAME}" .
chmod +x "${DIST_DIR}/${APP_NAME}" 2>/dev/null || true
echo -e "${GREEN}✓ Kompilasi sukses: ${DIST_DIR}/${APP_NAME} (${TARGET_ARCH})${NC}"

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
    [ -f "${SERVICE_NAME}.service" ] && cp -f "${SERVICE_NAME}.service" "${INSTALL_DIR}/"
fi

# 3. Beri permission eksekusi
chmod +x "${INSTALL_DIR}/bell-apps"

# 4. Pasang Systemd Service
if [ -f "${SERVICE_NAME}.service" ] || [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
    echo "Memasang systemd service..."
    [ -f "${SERVICE_NAME}.service" ] && cp -f "${SERVICE_NAME}.service" "/etc/systemd/system/${SERVICE_NAME}.service"
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

echo -e "\n${BLUE}[6/6] Mengompresi paket distribusi menjadi arsip .zip (${ZIP_NAME})...${NC}"

# Buat arsip ZIP riil (Deflate zip archive)
if command -v python >/dev/null 2>&1 || command -v python3 >/dev/null 2>&1; then
    PY_BIN="python"
    command -v python3 >/dev/null 2>&1 && PY_BIN="python3"
    $PY_BIN - << PYEOF
import zipfile, os

dist_dir = "${DIST_DIR}"
zip_name = "${ZIP_NAME}"
zip_path = os.path.join(dist_dir, zip_name)

with zipfile.ZipFile(zip_path, 'w', zipfile.ZIP_DEFLATED) as z:
    for root, dirs, files in os.walk(dist_dir):
        for f in files:
            if f.endswith('.zip') or f.endswith('.tar.gz'):
                continue
            full_path = os.path.join(root, f)
            rel_path = os.path.relpath(full_path, dist_dir)
            z.write(full_path, rel_path)
PYEOF
elif command -v zip >/dev/null 2>&1; then
    (cd "${DIST_DIR}" && zip -r -q "${ZIP_NAME}" . -x "*.zip" -x "*.tar.gz")
else
    powershell.exe -NoProfile -Command "Compress-Archive -Path '${DIST_DIR}/*' -DestinationPath '${DIST_DIR}/${ZIP_NAME}' -Force"
fi

echo -e "${GREEN}✓ Arsip ZIP berhasil dibuat: ${DIST_DIR}/${ZIP_NAME}${NC}"

echo -e "\n${GREEN}======================================================${NC}"
echo -e "${GREEN}   BUILD SELESAI DENGAN SUKSES!                       ${NC}"
echo -e "${GREEN}======================================================${NC}"
echo -e "Folder hasil : ${YELLOW}${DIST_DIR}/${NC}"
echo -e "File arsip   : ${YELLOW}${DIST_DIR}/${ZIP_NAME}${NC}"
echo -e ""
echo -e "Cara upload & deploy di VPS:"
echo -e "  1. Upload ${YELLOW}${DIST_DIR}/${ZIP_NAME}${NC} ke VPS:"
echo -e "     ${CYAN}scp ${DIST_DIR}/${ZIP_NAME} root@IP_VPS:/tmp/${NC}"
echo -e "  2. Di terminal VPS, ekstrak dan jalankan installer:"
echo -e "     ${CYAN}mkdir -p /var/www/bell-apps && unzip -o /tmp/${ZIP_NAME} -d /var/www/bell-apps${NC}"
echo -e "     ${CYAN}cd /var/www/bell-apps && bash install.sh${NC}"
echo -e ""
