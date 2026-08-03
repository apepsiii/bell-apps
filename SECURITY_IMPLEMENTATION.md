# Security Implementation Guide

**Project:** SMK NIBA Super Apps  
**Date:** 4 Agustus 2026  
**Version:** 2.1

---

## 📋 Overview

This document describes the security features implemented in Phase 4 & 5 of the audit fixes.

---

## 🛡️ Security Features Implemented

### 1. Security Headers Middleware

**File:** `internal/middleware/security.go`

**Headers Applied:**

| Header | Value | Purpose |
|--------|-------|---------|
| `Content-Security-Policy` | `default-src 'self'; script-src...` | Prevents XSS by restricting resource loading |
| `X-Frame-Options` | `DENY` | Prevents clickjacking attacks |
| `X-Content-Type-Options` | `nosniff` | Prevents MIME type sniffing |
| `X-XSS-Protection` | `1; mode=block` | Legacy XSS protection (browser-level) |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Controls referrer information leakage |
| `Permissions-Policy` | `geolocation=(), microphone=()...` | Disables unnecessary browser features |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Forces HTTPS (when TLS enabled) |

**Usage:**
```go
e.Use(appMiddleware.SecurityHeaders())
```

Applied globally to all routes in `internal/router/router.go:46`

---

### 2. Rate Limiting

**File:** `internal/middleware/security.go`

**Three Rate Limiters:**

#### a. API Rate Limiter
- **Limit:** 100 requests per minute per IP
- **Purpose:** General API abuse prevention
- **Applied to:** All `/api/*` endpoints (if needed)

#### b. Auth Rate Limiter
- **Limit:** 5 requests per minute per IP
- **Purpose:** Brute force protection on login
- **Applied to:**
  - `POST /api/login`
  - `POST /api/student/login`
  - `POST /api/operator/login`

#### c. Admin Rate Limiter
- **Limit:** 30 requests per minute per IP
- **Purpose:** Prevent admin operation abuse
- **Applied to:** All `/admin/*` endpoints

**Implementation Details:**
- In-memory storage with automatic cleanup
- Per-IP tracking using `c.RealIP()`
- Sliding window algorithm
- Automatic visitor cleanup every 1 minute

**Response on Rate Limit:**
```json
{
  "error": "Rate limit exceeded",
  "message": "Maksimal 100 request per 1m0s. Coba lagi nanti."
}
```
**Status Code:** `429 Too Many Requests`

---

### 3. CSRF Protection

**File:** `internal/middleware/csrf.go`

**How it Works:**

1. **Token Generation (GET requests):**
   - Generates 32-byte random token
   - Stores in memory with 24-hour expiration
   - Sets cookie `csrf_token`
   - Also available in context as `csrf_token` for templates

2. **Token Validation (POST/PUT/PATCH/DELETE):**
   - Checks header: `X-CSRF-Token`
   - Falls back to form field: `csrf_token`
   - Falls back to cookie: `csrf_token`
   - Validates token exists and not expired

3. **Automatic Cleanup:**
   - Expired tokens removed every hour
   - Prevents memory leaks

**Integration Examples:**

**HTML Form:**
```html
<form method="POST" action="/admin/student/add">
  <input type="hidden" name="csrf_token" value="{{.csrf_token}}">
  <input type="text" name="name" required>
  <button type="submit">Submit</button>
</form>
```

**JavaScript (Fetch API):**
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
  body: JSON.stringify({
    name: 'John Doe',
    // other fields...
  })
});
```

**jQuery:**
```javascript
$.ajax({
  url: '/api/admin/student/add',
  method: 'POST',
  headers: {
    'X-CSRF-Token': getCookie('csrf_token')
  },
  data: { name: 'John Doe' },
  success: function(data) { console.log(data); }
});

function getCookie(name) {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop().split(';').shift();
}
```

**CSRF Errors:**

| Error | Status | Message |
|-------|--------|---------|
| Token missing | 403 | Token CSRF tidak ditemukan |
| Token invalid | 403 | Token CSRF tidak valid |
| Token expired | 403 | Token CSRF sudah kadaluarsa, silakan refresh halaman |

---

## 🔧 Configuration

### Enabling/Disabling Features

**Router Configuration:** `internal/router/router.go`

```go
// Global security headers (recommended: always on)
e.Use(appMiddleware.SecurityHeaders())

// Rate limiting on auth endpoints
e.POST("/api/login", handler.Login(), appMiddleware.AuthRateLimiter())

// Rate limiting on admin routes
admin := e.Group("/admin", 
  appMiddleware.AdminAuth(), 
  appMiddleware.AdminRateLimiter())

// Optional: CSRF protection (not yet enabled globally)
// e.Use(appMiddleware.CSRF())
```

### Customizing Rate Limits

**Edit:** `internal/middleware/security.go`

```go
// Change API rate limit to 200/min
func APIRateLimiter() echo.MiddlewareFunc {
  limiter := NewRateLimiter(200, time.Minute)  // was 100
  return limiter.Middleware()
}

// Change auth rate limit to 10/min
func AuthRateLimiter() echo.MiddlewareFunc {
  limiter := NewRateLimiter(10, time.Minute)  // was 5
  return limiter.Middleware()
}
```

### Customizing CSP

**Edit:** `internal/middleware/security.go:15`

```go
// Add new trusted domain
csp := "default-src 'self'; " +
  "script-src 'self' 'unsafe-inline' https://trusted-cdn.com; " +
  // ... rest of policy
```

---

## 🧪 Testing

### 1. Test Security Headers

```bash
# Test security headers are present
curl -I http://localhost:8080/

# Expected headers:
# Content-Security-Policy: ...
# X-Frame-Options: DENY
# X-Content-Type-Options: nosniff
# X-XSS-Protection: 1; mode=block
# Referrer-Policy: strict-origin-when-cross-origin
# Permissions-Policy: ...
```

### 2. Test Rate Limiting

```bash
# Test auth rate limiter (should block after 5 attempts)
for i in {1..6}; do
  echo "Attempt $i:"
  curl -X POST http://localhost:8080/api/login \
    -d "username=test&password=wrong"
  echo ""
done

# Attempt 6 should return 429 Too Many Requests
```

### 3. Test CSRF Protection

```bash
# 1. Get CSRF token (GET request generates token)
CSRF_TOKEN=$(curl -c cookies.txt http://localhost:8080/admin | grep csrf_token)

# 2. Try POST without token (should fail)
curl -X POST http://localhost:8080/admin/student/add \
  -d "name=Test"
# Expected: 403 Forbidden

# 3. Try POST with token (should succeed)
curl -X POST http://localhost:8080/admin/student/add \
  -b cookies.txt \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -d "name=Test"
# Expected: 200 OK
```

### 4. Browser Testing

**Open Developer Tools → Network → Headers**

1. Visit any page: `http://localhost:8080`
2. Check response headers for security headers
3. Check cookies for `csrf_token`
4. Try submitting a form and verify CSRF token is sent

---

## 🚨 Important Notes

### CSP Considerations

The current CSP allows `'unsafe-inline'` and `'unsafe-eval'` for backward compatibility with existing frontend code. 

**To make it stricter:**
1. Remove inline scripts from HTML files
2. Add nonces to script tags
3. Update CSP to remove `'unsafe-inline'` and `'unsafe-eval'`

### Rate Limiting Storage

Current implementation uses **in-memory storage**:
- ✅ Fast and simple
- ⚠️ Resets on application restart
- ⚠️ Not shared across multiple instances

**For production with multiple instances:**
- Use Redis for shared rate limit storage
- Or use a reverse proxy (Nginx, Cloudflare) for rate limiting

### CSRF Token Storage

Current implementation uses **in-memory storage**:
- ✅ Fast and simple
- ⚠️ Tokens lost on application restart (users must refresh)
- ⚠️ Not shared across multiple instances

**For production with multiple instances:**
- Use Redis or database for token storage
- Or use signed tokens (no storage needed)

### HTTPS Requirement

Some security features work best with HTTPS:
- `Strict-Transport-Security` header only sent over HTTPS
- `Secure` cookie flag for CSRF token
- `SameSite=Strict` cookie attribute

**Setup HTTPS:**
- Use reverse proxy (Nginx) with Let's Encrypt certificate
- Or use Cloudflare in front of the application

---

## 📊 Performance Impact

| Feature | Impact | Notes |
|---------|--------|-------|
| Security Headers | Negligible | Simple header addition |
| Rate Limiting | Low | In-memory map lookup O(1) |
| CSRF Protection | Low | Token generation/validation |
| Memory Usage | +5-10 MB | For rate limiter + CSRF token storage |

**Estimated overhead:** < 1ms per request

---

## 🔒 Security Checklist

- [x] Security headers applied globally
- [x] Rate limiting on authentication endpoints
- [x] Rate limiting on admin operations
- [x] CSRF protection implemented
- [x] XSS prevention via CSP
- [x] Clickjacking prevention via X-Frame-Options
- [x] MIME sniffing prevention
- [x] Brute force protection (5 req/min on login)
- [x] API abuse prevention (100 req/min)
- [x] HTTPS enforcement ready (HSTS)
- [ ] CSRF enabled globally (optional - can be enabled per-route)
- [ ] Redis integration for multi-instance deployment
- [ ] Stricter CSP (remove unsafe-inline/unsafe-eval)

---

## 📚 References

- [OWASP Secure Headers Project](https://owasp.org/www-project-secure-headers/)
- [MDN Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
- [OWASP CSRF Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
- [OWASP Rate Limiting](https://cheatsheetseries.owasp.org/cheatsheets/Denial_of_Service_Cheat_Sheet.html)

---

**Last Updated:** 4 Agustus 2026 06:00 WIB
