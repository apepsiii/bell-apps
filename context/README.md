# SmartBell - School Management System

## System Requirements
- Go 1.25+
- SQLite (ModernC driver included)

## Local Development
1. Run `go run main.go`
2. Access `http://localhost:8080/login`
   - User: `admin`
   - Pass: `123456`

## Deployment to VPS (Linux/Ubuntu)

### Option 1: Binary Deployment (Easiest)
1. **Compile for Linux (from Windows):**
   Open terminal in project folder:
   ```powershell
   $Env:GOOS = "linux"; $Env:GOARCH = "amd64"; go build -o bell_linux main.go
   ```
   *This creates a `bell_linux` file.*

2. **Upload to VPS:**
   Use SCP or FileZilla to upload:
   - `bell_linux` (The binary)
   - `views/` (Folder)
   - `public/` (Folder)
   
   Target directory example: `/opt/smartbell`

3. **Run on VPS:**
   ```bash
   cd /opt/smartbell
   chmod +x bell_linux
   ./bell_linux
   ```

### Option 2: Git Deployment
1. Push this code to GitHub.
2. SSH into VPS.
3. Install Go: `sudo apt install golang`
4. Clone repo: `git clone https://github.com/username/repo.git`
5. Run: `go run main.go`

## Running in Background (Systemd)
Create service file: `/etc/systemd/system/bell.service`
```ini
[Unit]
Description=SmartBell Server
After=network.target

[Service]
User=root
WorkingDirectory=/opt/smartbell
ExecStart=/opt/smartbell/bell_linux
Restart=always

[Install]
WantedBy=multi-user.target
```
Enable: `sudo systemctl enable --now bell`
