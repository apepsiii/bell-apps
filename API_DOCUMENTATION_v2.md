# API Documentation - Audit Trail & Security

**Version:** 2.1  
**Last Updated:** 4 Agustus 2026  
**Changes:** Added audit trail, soft delete, validation, rate limiting, security headers, and CSRF protection

---

## 🔄 Breaking Changes

**None** - All changes are backward compatible.

---

## ✨ New Features

### 1. Soft Delete with Audit Trail
All delete operations now use soft delete and require:
- **Reason** (mandatory)
- **Authorization** (kepala_sekolah role)
- Logged to audit trail

### 2. Input Validation
- **Max points per transaction:** 100
- **Rate limiting:** 50 inputs per day per admin

### 3. Configurable Thresholds
All thresholds now stored in `point_config` table

---

## 📍 Endpoints

### POST /admin/v2/student/achievement
Add achievement point to student

**Request Body (FormData):**
```
student_id: string (required)
rule_id: string (optional - if empty, points required)
points: string (optional - overrides rule points)
description: string (optional)
recorded_by: string (default: "Admin")
academic_year: string (default: current year)
```

**Response 200 OK:**
```json
{
  "message": "Poin penghargaan berhasil dicatat",
  "points": 50
}
```

**Response 400 Bad Request:**
```json
{
  "error": "Maksimal poin per transaksi adalah 100"
}
```

**Response 429 Too Many Requests:**
```json
{
  "error": "Batas input harian tercapai",
  "message": "Anda sudah melakukan 50 input hari ini. Maksimal 50 per hari."
}
```

**Validations:**
- ✅ Max 100 points per transaction
- ✅ Rate limit 50 per day per admin
- ✅ Points must be positive
- ✅ Student ID required
- ✅ Audit log created automatically

---

### DELETE /admin/v2/achievement/:id
Soft delete achievement point (kepala_sekolah only)

**Request Body (FormData):**
```
reason: string (required, min 10 characters)
```

**Response 200 OK:**
```json
{
  "message": "Poin penghargaan berhasil dihapus"
}
```

**Response 400 Bad Request:**
```json
{
  "error": "Alasan penghapusan wajib diisi"
}
```

**Response 403 Forbidden:**
```json
{
  "error": "Hanya kepala sekolah yang dapat menghapus poin"
}
```

**Response 404 Not Found:**
```json
{
  "error": "Record tidak ditemukan atau sudah dihapus"
}
```

**Validations:**
- ✅ Authorization check (kepala_sekolah role)
- ✅ Reason required (min 10 chars)
- ✅ Soft delete (deleted_at, deleted_by, deletion_reason)
- ✅ Audit log created
- ✅ Cannot delete already deleted records

**Audit Log Entry:**
```json
{
  "action": "DELETE",
  "table_name": "student_achievement_points",
  "record_id": 123,
  "old_data": "{\"id\":123,\"student_id\":1,\"points\":50,...}",
  "new_data": null,
  "performed_by": "kepala_sekolah_username",
  "performed_at": "2026-08-04 05:30:00",
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "reason": "Input data salah"
}
```

---

### DELETE /admin/v2/violation/:id
Soft delete violation point (kepala_sekolah only)

**Same as DELETE /admin/v2/achievement/:id**

---

### GET /admin/v2/student/:id/points
Get student dual point profile

**Query Parameters:**
```
year: string (optional, default: current academic year)
```

**Response 200 OK:**
```json
{
  "student": {
    "id": 1,
    "name": "Ahmad Rizky",
    "class": "XII RPL 1",
    "nis": "12345",
    "photo": ""
  },
  "name": "Ahmad Rizky",
  "nis": "12345",
  "class_name": "XII RPL 1",
  "status": "active",
  "academic_year": "2026/2027",
  "achievement_points": 125,
  "violation_points": 15,
  "achievement_grade": "Sertifikat",
  "violation_level": "Aman",
  "class_rank": 3,
  "total_class_students": 35,
  "achievement_breakdown": [
    {
      "category": "Prestasi Akademik",
      "points": 50,
      "count": 5
    }
  ],
  "violation_breakdown": [
    {
      "category": "Terlambat",
      "points": 15,
      "count": 3
    }
  ],
  "achievement_history": [...],
  "violation_history": [...]
}
```

**Changes:**
- ✅ Now excludes soft-deleted records
- ✅ Added breakdown by category
- ✅ Added class ranking
- ✅ Added total class students

---

### GET /admin/v2/leaderboard
Get leaderboard sorted by achievement points

**Query Parameters:**
```
year: string (optional, default: current academic year)
```

**Response 200 OK:**
```json
[
  {
    "id": 1,
    "name": "Ahmad Rizky",
    "class_name": "XII RPL 1",
    "achievement_points": 125,
    "violation_points": 15,
    "achievement_grade": "Sertifikat",
    "violation_level": "Aman",
    "rank": 1
  }
]
```

**Changes:**
- ✅ Now excludes soft-deleted records
- ✅ Accurate ranking based on active points only

---

### GET /admin/v2/class-summary
Get summary of dual-track points per class

**Query Parameters:**
```
year: string (optional, default: current academic year)
```

**Response 200 OK:**
```json
[
  {
    "id": 1,
    "name": "XII RPL 1",
    "student_count": 35,
    "total_achievement": 3500,
    "total_violation": 450
  }
]
```

**Changes:**
- ✅ Now excludes soft-deleted records
- ✅ Accurate totals based on active points only

---

## 🗄️ Database Schema Changes

### New Tables

#### point_audit_log
```sql
CREATE TABLE point_audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    action TEXT NOT NULL,              -- INSERT, UPDATE, DELETE
    table_name TEXT NOT NULL,          -- student_achievement_points, etc.
    record_id INTEGER NOT NULL,
    old_data TEXT,                     -- JSON before change
    new_data TEXT,                     -- JSON after change
    performed_by TEXT NOT NULL,        -- Username
    performed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    reason TEXT
);

-- Indexes
CREATE INDEX idx_audit_log_table_record ON point_audit_log(table_name, record_id);
CREATE INDEX idx_audit_log_performed_by ON point_audit_log(performed_by);
CREATE INDEX idx_audit_log_performed_at ON point_audit_log(performed_at);
```

#### point_config
```sql
CREATE TABLE point_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    config_key TEXT UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    description TEXT,
    updated_by TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Default values
INSERT INTO point_config (config_key, config_value, description) VALUES
    ('MAX_POINTS_PER_TRANSACTION', '100', 'Maksimal poin per transaksi'),
    ('RATE_LIMIT_PER_DAY', '50', 'Maksimal input per hari per admin'),
    ('APPROVAL_THRESHOLD', '50', 'Poin lebih dari ini perlu approval'),
    ('SP1_THRESHOLD', '25', 'Threshold untuk SP1'),
    ('SP2_THRESHOLD', '51', 'Threshold untuk SP2'),
    ('SP3_THRESHOLD', '76', 'Threshold untuk SP3'),
    ('CERT_THRESHOLD', '100', 'Threshold untuk sertifikat'),
    ('REWARD_THRESHOLD', '126', 'Threshold untuk hadiah'),
    ('WALUYA_THRESHOLD', '151', 'Threshold untuk Waluya Utama');
```

#### pending_achievement_points
```sql
CREATE TABLE pending_achievement_points (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    rule_id INTEGER,
    points INTEGER NOT NULL,
    description TEXT,
    requested_by TEXT NOT NULL,
    requested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    approved_by TEXT,
    approved_at DATETIME,
    rejected_by TEXT,
    rejected_at DATETIME,
    status TEXT DEFAULT 'pending',     -- pending, approved, rejected
    rejection_reason TEXT,
    FOREIGN KEY (student_id) REFERENCES students(id)
);
```

### Modified Tables

#### student_achievement_points
```sql
ALTER TABLE student_achievement_points ADD COLUMN deleted_at DATETIME;
ALTER TABLE student_achievement_points ADD COLUMN deleted_by TEXT;
ALTER TABLE student_achievement_points ADD COLUMN deletion_reason TEXT;
```

#### student_violation_points
```sql
ALTER TABLE student_violation_points ADD COLUMN deleted_at DATETIME;
ALTER TABLE student_violation_points ADD COLUMN deleted_by TEXT;
ALTER TABLE student_violation_points ADD COLUMN deletion_reason TEXT;
```

---

## 🔒 Authorization

### Role Hierarchy
```
superadmin > kepala_sekolah > admin > guru
```

### Permissions

| Action | superadmin | kepala_sekolah | admin | guru |
|--------|-----------|----------------|-------|------|
| Add points | ✅ | ✅ | ✅ | ✅ |
| Delete points | ✅ | ✅ | ❌ | ❌ |
| View audit log | ✅ | ✅ | ❌ | ❌ |
| Edit config | ✅ | ❌ | ❌ | ❌ |
| Hard delete | ✅ | ❌ | ❌ | ❌ |

**Note:** Current implementation returns true for all roles for backward compatibility. Implement proper role checking in production.

---

## 📊 Rate Limiting

### Rules
- **50 inputs per day** per admin (configurable via point_config)
- Counter resets at midnight
- Counted per `recorded_by` field
- Only counts successful inserts
- Excludes soft-deleted records

### Error Response
```json
{
  "error": "Batas input harian tercapai",
  "message": "Anda sudah melakukan 50 input hari ini. Maksimal 50 per hari."
}
```

**HTTP Status:** 429 Too Many Requests

---

## 🧪 Testing

### cURL Examples

**1. Add Point (Valid):**
```bash
curl -X POST http://localhost:8080/admin/v2/student/achievement \
  -d "student_id=1" \
  -d "rule_id=1" \
  -d "points=50" \
  -d "description=Test" \
  -d "recorded_by=admin"
```

**2. Add Point (Over Limit):**
```bash
curl -X POST http://localhost:8080/admin/v2/student/achievement \
  -d "student_id=1" \
  -d "points=150"
# Expected: {"error":"Maksimal poin per transaksi adalah 100"}
```

**3. Delete Point (With Reason):**
```bash
curl -X DELETE http://localhost:8080/admin/v2/achievement/1 \
  -d "reason=Input data salah, siswa tidak hadir"
```

**4. Delete Point (No Reason):**
```bash
curl -X DELETE http://localhost:8080/admin/v2/achievement/1
# Expected: {"error":"Alasan penghapusan wajib diisi"}
```

**5. Rate Limit Test:**
```bash
for i in {1..51}; do
  curl -X POST http://localhost:8080/admin/v2/student/achievement \
    -d "student_id=1" \
    -d "rule_id=1" \
    -d "recorded_by=test_admin"
  echo "Request $i done"
done
# After 50: {"error":"Batas input harian tercapai"}
```

---

## 📝 Migration Guide

### For Existing Deployments

1. **Backup database:**
```bash
sqlite3 database.db ".backup database_backup_$(date +%Y%m%d).db"
```

2. **Deploy new binary:**
```bash
./smartbell.exe
```

3. **Migration runs automatically on startup**

4. **Verify migration:**
```bash
sqlite3 database.db "SELECT name FROM sqlite_master WHERE type='table';"
# Should show: point_audit_log, point_config, pending_achievement_points
```

### Rollback
```bash
# Stop application
# Restore backup
mv database_backup.db database.db
# Deploy old binary
```

---

## 🔒 Security Features (Phase 4 & 5)

### 1. Security Headers
All responses now include:

**Content Security Policy (CSP):**
- Restricts resource loading to trusted sources
- Prevents XSS attacks
- Blocks unsafe inline scripts (with exceptions for legacy code)

**HTTP Headers Applied:**
```
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net https://unpkg.com; ...
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=(self), payment=()
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload (HTTPS only)
```

### 2. Rate Limiting

**Global API Rate Limit:**
- 100 requests per minute per IP
- Applied to all `/api/*` endpoints

**Authentication Rate Limit:**
- 5 requests per minute per IP
- Applied to login endpoints:
  - `POST /api/login`
  - `POST /api/student/login`
  - `POST /api/operator/login`
- Prevents brute force attacks

**Admin Operations Rate Limit:**
- 30 requests per minute per IP
- Applied to all `/admin/*` endpoints
- Prevents abuse of administrative functions

**Rate Limit Response:**
```json
{
  "error": "Rate limit exceeded",
  "message": "Maksimal 100 request per 1m0s. Coba lagi nanti."
}
```
**Status Code:** `429 Too Many Requests`

### 3. CSRF Protection

**Implementation:**
- Token-based CSRF protection for all state-changing operations
- Automatic token generation for GET requests
- Token validation for POST/PUT/PATCH/DELETE requests

**Token Delivery:**
- Cookie: `csrf_token` (24 hour expiry)
- Also available in context for template rendering

**Token Submission (choose one):**
1. **Header:** `X-CSRF-Token: <token>`
2. **Form field:** `csrf_token=<token>`
3. **Cookie:** Automatically checked if not in header/form

**CSRF Error Responses:**

Missing token:
```json
{
  "error": "CSRF token missing",
  "message": "Token CSRF tidak ditemukan"
}
```

Invalid token:
```json
{
  "error": "Invalid CSRF token",
  "message": "Token CSRF tidak valid"
}
```

Expired token:
```json
{
  "error": "CSRF token expired",
  "message": "Token CSRF sudah kadaluarsa, silakan refresh halaman"
}
```
**Status Code:** `403 Forbidden`

**Usage Example:**

HTML Form:
```html
<form method="POST" action="/admin/student/add">
  <input type="hidden" name="csrf_token" value="{{.csrf_token}}">
  <!-- other fields -->
</form>
```

JavaScript (Fetch API):
```javascript
// Get token from cookie
const csrfToken = document.cookie
  .split('; ')
  .find(row => row.startsWith('csrf_token='))
  ?.split('=')[1];

fetch('/api/admin/student/add', {
  method: 'POST',
  headers: {
    'X-CSRF-Token': csrfToken,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({...})
});
```

---

## 🐛 Known Issues

1. **CheckAdminRole** currently returns true by default
   - **Workaround:** Implement proper role from session
   
2. **Rate limit uses DATE('now')** - SQLite specific
   - **Workaround:** Adjust for MySQL if needed

3. **User context not fully implemented**
   - **Workaround:** Defaults to "Admin"

4. **CSRF Protection** - Optional for backward compatibility
   - Currently not enforced globally
   - Can be enabled per-route or globally in router.go

---

## 📞 Support

- **Documentation:** AUDIT_SISTEM_POINT.md
- **Progress:** AUDIT_FIX_PROGRESS.md
- **Report:** SPRINT1_COMPLETION_REPORT.md

---

**Last Updated:** 4 Agustus 2026 06:00 WIB
