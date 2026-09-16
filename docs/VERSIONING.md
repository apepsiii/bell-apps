# NIBA SuperApps - Versioning System

**Version:** 2.1.0  
**Release Date:** 2026-09-16

## Overview

NIBA SuperApps sekarang memiliki sistem versioning lengkap yang menampilkan informasi versi di UI, API, dan logs.

## Fitur Versioning

### 1. Version Variables (version.go)

```go
Version   = "2.1.0"        // Semantic version
BuildDate = "unknown"       // Build timestamp (UTC)
GitCommit = "unknown"       // Git commit hash (short)
AppName   = "NIBA SuperApps"
```

### 2. Build-time Injection

Build script (`build.sh`) inject version ke binary via ldflags:

```bash
VERSION="2.1.0"
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD)
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}"
```

### 3. Environment Override

Version bisa di-override via environment variables:

```bash
export APP_VERSION=2.1.0
export BUILD_DATE=2026-09-16T16:40:00Z
export GIT_COMMIT=abc1234
export APP_NAME="NIBA SuperApps"
```

### 4. UI Display

**Login Page Footer:**
```
NIBA SuperApps v2.1.0
Build 2026-09-16T16:40:00Z · abc1234
```

**Admin Dashboard:**
- User dropdown: `v2.1.0`
- Footer: `NIBA SuperApps v2.1.0 · Build 2026-09-16T16:40:00Z · abc1234`

### 5. API Endpoint

**GET** `/api/version`

Response:
```json
{
  "version": "2.1.0",
  "build_date": "2026-09-16T16:40:00Z",
  "git_commit": "abc1234",
  "app_name": "NIBA SuperApps"
}
```

### 6. Startup Logs

```
INFO starting NIBA SuperApps version=v2.1.0 build=2026-09-16T16:40:00Z commit=abc1234
```

## Build Instructions

### Lokal (Windows)

```powershell
$VERSION="2.1.0"
$BUILD_DATE=(Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
$GIT_COMMIT=(git rev-parse --short HEAD)
$LDFLAGS="-s -w -X main.Version=$VERSION -X main.BuildDate=$BUILD_DATE -X main.GitCommit=$GIT_COMMIT"
$env:GOOS="linux"
$env:GOARCH="arm64"
$env:CGO_ENABLED="0"
go build -ldflags="$LDFLAGS" -o dist/smartbell_linux_arm64 .
```

### Linux/Mac (via build.sh)

```bash
VERSION=2.1.0 sh build.sh
# atau
sh build.sh  # default version dari script
```

Build script otomatis:
- ✅ Inject version dari `VERSION` env atau default `2.1.0`
- ✅ Generate `BUILD_DATE` dari waktu build
- ✅ Extract `GIT_COMMIT` dari git
- ✅ Build untuk amd64 + arm64
- ✅ Copy `setup.sh` ke dist/

## Version Management

### Semantic Versioning

Format: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes, incompatible API changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Current Version History

| Version | Date | Changes |
|---------|------|---------|
| 2.1.0 | 2026-09-16 | Gowa WhatsApp Gateway integration, versioning system |
| 2.0.0 | 2026-09-13 | Composite scoring system (attendance index, achievements, violations, leaderboard) |
| 1.3.0 | 2026-07-20 | Manual attendance improvements |

### Release Process

1. **Update version.go**
   ```go
   Version = "2.2.0"  // Bump version
   ```

2. **Commit & Tag**
   ```bash
   git add version.go
   git commit -m "chore: bump version to 2.2.0"
   git tag v2.2.0
   git push origin main --tags
   ```

3. **Build Release**
   ```bash
   VERSION=2.2.0 sh build.sh
   ```

4. **Deploy**
   ```bash
   scp dist/smartbell_linux_arm64 user@vps:/opt/nibasuperappsv2/
   ssh user@vps "cd /opt/nibasuperappsv2 && sudo bash setup.sh"
   ```

## Configuration Files

### .env.example

```bash
APP_VERSION=2.1.0
APP_NAME=NIBA SuperApps
BUILD_DATE=2026-09-16
GIT_COMMIT=unknown
```

### config.yaml

Version tidak disimpan di config.yaml — source of truth adalah:
1. Build-time ldflags (highest priority)
2. Environment variables
3. version.go defaults

## Template Access

Semua template otomatis mendapat variable version:

```html
{{.AppVersion}}     <!-- v2.1.0 -->
{{.AppName}}        <!-- NIBA SuperApps -->
{{.BuildDate}}      <!-- 2026-09-16T16:40:00Z -->
{{.GitCommit}}      <!-- abc1234 -->
{{.FullVersion}}    <!-- NIBA SuperApps v2.1.0 (build 2026-09-16T16:40:00Z) -->
{{.VersionInfo}}    <!-- map[string]string dengan semua info -->
```

## Troubleshooting

### Version tidak update setelah build

**Penyebab:** Binary tidak di-rebuild atau ldflags tidak dijalankan

**Solusi:**
```bash
# Clean build
rm dist/*
VERSION=2.1.0 sh build.sh

# Verify
./dist/smartbell_linux_arm64 --help 2>&1 | grep version
```

### Version masih "v1.3.0" di UI

**Penyebab:** Template cache atau browser cache

**Solusi:**
1. Hard refresh browser (Ctrl+Shift+R)
2. Restart service: `sudo systemctl restart nibasuperappsv2`
3. Check logs: `journalctl -u nibasuperappsv2 -n 50`

### Git commit "unknown"

**Penyebab:** Build di luar git repo atau git tidak installed

**Solusi:**
```bash
# Build di dalam git repo
cd /path/to/bell-apps
git status  # verify repo
sh build.sh
```

## Files Modified

- `version.go` - Version variables & functions
- `main.go` - VersionHandler + template injection + startup log
- `internal/router/router.go` - Route `/api/version`
- `internal/config/config.go` - AppVersion sync
- `views/login.html` - Footer version display
- `views/admin.html` - User dropdown + footer version
- `build.sh` - Ldflags injection
- `.env.example` - Version env vars

## Future Enhancements

- [ ] Changelog generator otomatis dari git commits
- [ ] Version compare untuk migration compatibility check
- [ ] Auto-update notification di admin dashboard
- [ ] Release notes display di UI
- [ ] Rollback ke version sebelumnya via setup.sh
