# Progress: UNPIX Design Implementation (Student Portal Redesign)

## Tanggal: 2026-07-23

## Goal
Mengimplementasikan desain UNPIX (context/studentlogin.html + context/home.html) ke halaman student portal SMK NIBA Super Apps, mengubah arsitektur dari multi-page (separate HTML files) menjadi SPA (Single Page Application).

## Done
1. **`views/student/login.html`** — Disedain ulang berdasarkan context/studentlogin.html:
   - Mobile container dengan phone frame (414px, rounded di desktop)
   - Hero section: gradient biru (#0B57C2 → #1565D8), decorative stars (clip-path), SVG illustration (dua orang high-five), wave SVG di bawah
   - Branding: "SMK NIBA" + "Super Apps" (bukan UNPIX)
   - Input NIS/NISN (bukan HP/Email) dengan icon `ph-identification-card`
   - Input PIN (bukan Password) dengan icon `ph-lock-key`, toggle visibility (`ph-eye`/`ph-eye-slash`), `inputmode="numeric" maxlength="8"`
   - Error box dengan icon `ph-warning-circle`
   - Tombol Masuk dengan spinner loading state
   - Hint "PIN default: 123456" di bawah form
   - Dihapus: biometric button, forgot password link, register link
   - API: `fetch('/api/student/login')` → redirect ke `/student/app` (bukan `/student/dashboard`)
   - Phosphor Icons CDN (`@phosphor-icons/web`), bukan Lucide

2. **`views/student/app.html`** — SPA baru berdasarkan context/home.html, menggabungkan 4 halaman jadi 1:
   - **Mobile container** dengan phone frame, same blue scheme
   - **5 view sections** dengan `switchTab()` + fade-in animation:
     - **Home (Beranda)**: top header biru gradient (avatar, "SMK NIBA" branding, notif bell), Total Poin display dengan toggle hide/show, menu grid 4x2 (Presensi aktif, 7 lainnya opacity-40 disabled "segera hadir"), Kehadiran Hari Ini widget (jam masuk/pulang, status badge), Riwayat Poin Terakhir
     - **Presensi**: top header kecil, calendar grid dengan month navigation (prev/next), status dots (hijau=hadir, kuning=izin/sakit/telat, merah=alpa), legend, detail kehadiran harian (klik tanggal → jam masuk/pulang)
     - **QR Siswa**: full blue background header, QR card overlap (-mt-60), student photo + name + NIS + status, QR image dengan scan-line animation, tombol Unduh + Bagikan, auto-refresh hint
     - **Info (Pengumuman)**: top header kecil, list pengumuman dari dashboard API dengan icon megaphone
     - **Profil**: header putih dengan photo + name + class, info rows (NIS, birthday, wali, HP wali), form Ganti PIN (old/new/confirm), tombol Keluar (logout)
   - **Bottom nav** dengan floating QR button di tengah (qris-btn-container, gradient biru), 4 nav items: Beranda, Presensi, QR (center), Info, Profil
   - **JS logic**:
     - `loadDashboard()` → fetch `/api/student/dashboard`, populate home + profil + info
     - `loadCalendar()` → fetch `/api/student/calendar?month=X&year=Y`, render grid + detail
     - `loadQRCard()` → fetch `/api/student/qrcard`, render QR image + student info
     - `switchTab(tabName)` → toggle views, update nav active states, lazy-load data per tab
     - Initial tab determined dari `window.location.pathname` (e.g. `/student/presensi` → presensi tab)
     - PIN change → `PUT /api/student/pin` dengan JSON body
     - Logout → `POST /api/student/logout` → redirect `/student/login`
     - Toast notification untuk default PIN reminder
   - Tailwind config: brand.blue `#0B57C2`, brand.dark `#084091`, brand.light `#4EA6E9`, brand.bg `#F3F8FF`, ui.success/danger/warning
   - Custom shadows: soft, nav, icon, qr
   - Skeleton loading states untuk QR card

3. **`main.go` routes updated**:
   - Line 3362: Login redirect `/student/dashboard` → `/student/app`
   - Line 3376-3382: 
     - `/student` → redirect ke `/student/app`
     - `/student/app` → serve `app.html` (NEW)
     - `/student/dashboard` → serve `app.html` (was dashboard.html)
     - `/student/presensi` → serve `app.html` (was attendance.html)
     - `/student/qr` → serve `app.html` (was qrid.html)
     - `/student/profil` → serve `app.html` (was profile.html)
   - `//go:embed` directive (line 77) sudah cover `views/student/*.html` — app.html otomatis ter-embed

## In Progress
- Build & verify (belum dijalankan)

## Blocked
- (none)

## Pending
- Run `go build` untuk verify compilation
- Run `go vet ./internal/...` 
- Run student portal tests (`go test -run Student`)
- Test `student_portal_test.go` line 302: masih expect redirect ke `/student/login` — ini untuk middleware test, tidak perlu diubah (middleware redirect tidak berubah, hanya page redirect yang berubah)
- Old files (`dashboard.html`, `attendance.html`, `qrid.html`, `profile.html`) masih ada di `views/student/` — bisa dihapus nanti atau dibiarkan (tidak ter-serve tapi tetap ter-embed)
- Verify: login.html redirect ke `/student/app` — test di student_portal_test.go tidak test halaman login redirect, hanya API redirect

## Key Decisions
- **SPA approach**: Semua student pages jadi 1 file `app.html` dengan tab switching via JS — lebih native app-like, no page reloads, bottom nav always visible
- **Lazy loading**: Dashboard data load saat page load, calendar & QR load saat tab pertama kali dibuka
- **URL-based initial tab**: `/student/presensi` → presensi tab, `/student/qr` → qr tab, dst. JS baca `window.location.pathname`
- **Old views preserved**: dashboard.html, attendance.html, qrid.html, profile.html tidak dihapus (masih ter-embed tapi tidak di-route)
- **Phosphor Icons** menggantikan Lucide untuk konsistensi dengan desain UNPIX
- **Blue color scheme** (#0B57C2) menggantikan Islamic emerald (#0F5132) — mengikuti desain yang user buat

## Relevant Files
- `views/student/login.html` — Login page (UNPIX design, rewritten)
- `views/student/app.html` — SPA app (NEW, combines all 4 views)
- `views/student/dashboard.html` — Old dashboard (unused, preserved)
- `views/student/attendance.html` — Old attendance (unused, preserved)
- `views/student/qrid.html` — Old QR ID (unused, preserved)
- `views/student/profile.html` — Old profile (unused, preserved)
- `main.go` — Routes updated (lines 3356-3382)
- `context/studentlogin.html` — Source design untuk login
- `context/home.html` — Source design untuk app SPA

## Next Steps
1. Run `go build` — verify compilation
2. Run `go vet ./internal/...` — verify no issues
3. Run `go test -run Student ./...` — verify tests pass
4. Manual test: login → app → switch tabs → verify data loads
5. Pertimbangkan hapus old view files (dashboard.html, attendance.html, qrid.html, profile.html) untuk mengurangi binary size
