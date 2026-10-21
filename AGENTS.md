# AI Agents & Maintainer Instructions — SMK NIBA Super Apps (`bell-apps`)

File ini adalah instruksi operasional untuk AI Coding Agent (Gemini, Claude, Copilot, ChatGPT) dan pengembang dalam memelihara, memodifikasi, dan mengeksekusi redesain sistem **SMK NIBA Super Apps**.

---

## 📌 Pedoman Umum Proyek

- **Bahasa & Framework Backend:** Go (Echo v4 / `net/http`), SQLite (`modernc.org/sqlite`), Go Embed (`viewsFS`).
- **Frontend & Rendering:** Server-Side Rendered (SSR) menggunakan Go `html/template` (`views/*.html`, `views/admin/*.html`).
- **Styling:** Tailwind CSS (CDN) + Semantic CSS Tokens & Komponen Semantik (berbasis filosofi [shadcn-html](https://github.com/codylindley/shadcn-html)).
- **Ikon:** Lucide Icons (`<i data-lucide="..."></i>` dan inisialisasi `lucide.createIcons()`).

---

## 🎨 Pedoman Redesain UI (Gaya shadcn)

> ⚠️ **PENTING:** Seluruh pengerjaan antarmuka UI wajib mengacu pada **[`REDESIGN.md`](./REDESIGN.md)**.

### 1. Aturan Desain & Palet Warna
- **TIDAK LAGI MENGGUNAKAN** tema lama Stisla: dilarang menambahkan header ungu (`.navbar-bg`), tombol bayangan ungu tebal, atau layout ala template Stisla lawas.
- **GUNAKAN** palet netral modern (Zinc / Slate):
  - Canvas / Background: `bg-[#fafafa]` (Zinc-50) / dark: `bg-[#09090b]` (Zinc-950).
  - Cards & Containers: `bg-white border border-zinc-200/80 shadow-xs dark:bg-zinc-900/60 dark:border-zinc-800`.
  - Primary Action Buttons: `bg-zinc-900 text-white hover:bg-zinc-800 dark:bg-zinc-50 dark:text-zinc-900`.
  - Destructive Actions: `bg-rose-600 text-white hover:bg-rose-700`.
  - Secondary / Outline: `border border-zinc-200 bg-white hover:bg-zinc-100 dark:border-zinc-800 dark:bg-zinc-900 dark:hover:bg-zinc-800`.
  - Focus Ring: `focus:ring-2 focus:ring-zinc-950/20 focus:border-zinc-900 dark:focus:ring-zinc-300/20`.

### 2. Aturan Aksesibilitas & Dark Mode
- Setiap elemen kartu, dropdown, modal, tabel, dan form input **harus mendukung Dark Mode** secara eksplisit (`dark:...`).
- Dukung penyesuaian tema `data-theme="dark"` / `data-theme="light"` yang tersimpan di `localStorage`.

### 3. Integritas Fungsionalitas & JavaScript (CRITICAL)
- **DILARANG MENGHAPUS / MENGUBAH SEMBARANGAN**:
  - ID elemen HTML (contoh: `#search-student`, `#tab-btn-students`, `#bulk-action-bar`, dsb) karena terhubung langsung dengan fungsi JavaScript di bawah template atau file eksternal.
  - Atribut form handler (`onsubmit="..."`, `onclick="..."`, `onchange="..."`).
  - Elemen modal/dialog yang dibuka via JavaScript atau SweetAlert2.
- **CSRF Token:** Semua request mutating (`POST`, `PUT`, `DELETE`) wajib menyertakan CSRF header via `window.getCsrfToken()`. Jangan pernah merusak fungsi interceptor fetch di `layout_head.html`.

### 4. Pedoman Portal Siswa (`views/student/app.html` & `views/student/login.html`)
- **Prinsip Mobile-First & Desktop Centering:**
  - Tampilan dirancang native mobile, namun pada desktop harus selalu rapi berada di tengah dalam `.mobile-container` (max-width 480px, centered dengan border/shadow halus).
  - Pastikan seluruh view section (Home, Presensi, QR, Poin, English, Notifikasi, Cuaca, Profil, Belajar/LMS) berada di dalam `div.mobile-container > main#main-content`. Jangan pernah membiarkan penutup `</div>` ganda yang memotong container lebih awal.
- **Palet Warna Electric Violet FinTech:**
  - Kartu hero & header menggunakan gradient gelap elegan: `linear-gradient(135deg, #1A0B2E 0%, #2E1065 60%, #4C1D95 100%)`.
  - Teks di atas kartu gelap **wajib kontras terang**: `text-white` untuk jam/angka utama, `text-purple-200` atau `text-purple-300` untuk label dan unit (WIB).
- **Hero Attendance Card (Presensi Gerbang):**
  - Menggantikan quick actions bar lama di posisi floating hero (`-mt-7 relative z-20`).
  - DOM IDs wajib dijaga untuk binding JavaScript: `#todayDateBadge`, `#todayInStatus`, `#todayInTime`, `#todayInLabel`, `#todayOutTime`, `#todayOutLabel`, `#holidayNoteHome`.
  - Fungsi `renderTodayAttendance(today)` di JS **dilarang** meng-override teks jam menjadi gelap (`text-text-main` / `text-text-muted`). Selalu gunakan `text-white` dan badge semi-transparan `bg-emerald-500/25 text-emerald-300 border border-emerald-500/30`.
- **Modul Doa Harian & Dzikir:**
  - Endpoint backend terpusat di `internal/handler/doa.go` (`GET /api/student/doa`) dengan cache in-memory 1 jam.
  - UI modal berada di `#doaModal`, dipanggil via `openDoaModal()`.

### 5. Pedoman Modul LMS & Portal Guru (`views/teacher/*` & `views/student/app.html`)
- **Portal Guru & Kelas KBM:**
  - File target: `views/teacher/login.html`, `views/teacher/dashboard.html`, `views/teacher/class_detail.html`.
  - Backend handlers: `internal/handler/teacher_portal.go`, `internal/handler/teacher_kbm.go`, `internal/handler/lms_materials_assignments.go`, `internal/handler/lms_quiz.go`.
  - Autentikasi guru menggunakan cookie HTTP-only `teacher_session` dan pencocokan sesi pada tabel `teacher_sessions`.
  - Setiap request mutating (`POST`, `PUT`, `DELETE`) dari antarmuka guru wajib menyertakan CSRF header via interceptor `fetch`.
- **Sistem Banner Mata Pelajaran & Kelas KBM:**
  - File banner vektor statis disimpan di `public/assets/banners/*.svg` (8 preset: `math`, `tech`, `language`, `science`, `business`, `art`, `religion`, `general`).
  - Logika fallback otomatis diatur oleh fungsi `DefaultBannerForSubject(category, code, name string) string` di `internal/handler/lms_admin.go`.
  - Selalu sertakan `<img src="${banner_url || '/assets/banners/general.svg'}">` dan gradien gelap kontras tinggi saat merender kartu mapel/kelas di Guru dan Siswa.
- **Penyimpanan Berkas Tugas & Materi:**
  - Pengelolaan berkas upload wajib melalui `StorageService` (`internal/service/storage.go`), mendukung penyimpanan berkas lokal (`public/uploads`) maupun Cloudflare R2 / AWS S3.

### 6. Pedoman Halaman Error Sistem (404 & 403)
- File target: `views/404.html` dan `views/403.html`.
- Gaya wajib mengikuti standar shadcn UI netral (Zinc / Slate):
  - Canvas: `bg-[#fafafa]` / `dark:bg-[#09090b]` dengan ambient dot grid `bg-[radial-gradient(#e4e4e7_1px,transparent_1px)]`.
  - Status badge mono: `ERROR 404 • NOT FOUND` / `ERROR 403 • ACCESS FORBIDDEN`.
  - Dukungan tombol aksi ganda: *Kembali ke Beranda* (`/`) dan *Sebelumnya* (`history.back()`).
  - Penanganan error terpusat di `internal/router/router.go` pada `e.HTTPErrorHandler`.

---

## 📋 Alur Kerja Migrasi Bertahap (Roadmap Tracking)

Setiap agen yang mengerjakan tugas redesain **WAJIB** membaca status terakhir di **[`REDESIGN.md`](./REDESIGN.md)** dan memperbarui checklist:

1. **Fase 1: Token & Shell Global (SELESAI ✅)**
   - File target: `views/admin/layout_head.html`, `views/admin/layout_sidebar.html`, `views/admin/layout_topbar.html`, `views/admin/layout_footer.html`.
2. **Fase 2: Standarisasi Primitif Komponen (SELESAI ✅)**
   - Komponen standar: `.btn`, `.card`, input form, badges, tabel.
3. **Fase 3: Pilot Page (Low Risk) (SELESAI ✅)**
   - File target: `views/admin/settings.html` dan `views/admin/audio.html`.
4. **Fase 4: Core Modul (Data-Heavy) (SELESAI ✅)**
   - File target: `dashboard.html`, `academics.html`, `attendance.html`, `schedules.html`, `points.html`, `reports.html`, `student_detail.html`.
5. **Fase 5: Kiosk & Public Views (SELESAI ✅)**
   - File target: `scan.html`, `scan_sholat.html`, `attendance_face.html`, `face_register.html`, `login.html`, `idcard_single.html`.
6. **Fase 6: Sarpras & Asset Management (SELESAI ✅)**
   - File target: `views/admin/sarpras.html`, `internal/handler/sarpras_page.go`, `internal/handler/asset_excel.go`.
7. **Fase 7: Portal Siswa (Mobile-First Redesign) (SELESAI ✅)**
   - File target: `views/student/app.html`, `views/student/login.html`, `internal/handler/doa.go`.
8. **Fase 8: Modul LMS & Portal Guru (SELESAI ✅)**
   - File target: `views/teacher/login.html`, `views/teacher/dashboard.html`, `views/teacher/class_detail.html`, `internal/handler/teacher_portal.go`, `internal/handler/teacher_kbm.go`, `internal/handler/lms_materials_assignments.go`, `internal/handler/lms_quiz.go`, `internal/models/lms.go`, `internal/service/storage.go`.
9. **Fase 9: Halaman Error & Public Pages (SELESAI ✅)**
   - File target: `views/404.html`, `views/403.html`, `views/index.html`.

---

## 🛠️ Perintah Berguna untuk Testing & Build

- **Jalankan Aplikasi Lokal:**
  ```powershell
  go run main.go
  ```
  *(Aplikasi berjalan di port yang ditentukan di config/terminal, default `http://localhost:8080`)*

- **Cek Kompilasi Go Template & Syntax:**
  ```powershell
  go vet ./...
  go build -o nul .
  ```

- **Jalankan Test Suite:**
  ```powershell
  go test ./tests/...
  ```

---

## 💡 Instruksi Khusus untuk AI Agent Saat Melakukan Edit
1. Lakukan perubahan per bagian/modul kecil secara rapi.
2. Selalu cek kembali bahwa tag penutup `{{end}}` dan Go template blocks (`{{define "..."}}`, `{{template "..." .}}`) tidak rusak.
3. Setelah mengedit tampilan, periksa apakah ada JavaScript inline yang terdampak class name barunya.
4. Jangan menghapus komentar fungsional yang sudah ada di template.

