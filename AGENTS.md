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

---

## 📋 Alur Kerja Migrasi Bertahap (Roadmap Tracking)

Setiap agen yang mengerjakan tugas redesain **WAJIB** membaca status terakhir di **[`REDESIGN.md`](./REDESIGN.md)** dan memperbarui checklist:

1. **Fase 1: Token & Shell Global**
   - File target: `views/admin/layout_head.html`, `views/admin/layout_sidebar.html`, `views/admin/layout_topbar.html`, `views/admin/layout_footer.html`.
2. **Fase 2: Standarisasi Primitif Komponen**
   - Komponen standar: `.btn`, `.card`, input form, badges, tabel.
3. **Fase 3: Pilot Page (Low Risk)**
   - File target: `views/admin/settings.html` dan `views/admin/audio.html`.
4. **Fase 4: Core Modul (Data-Heavy)**
   - File target: `dashboard.html`, `academics.html`, `attendance.html`, `schedules.html`, `points.html`, `reports.html`, `student_detail.html`.
5. **Fase 5: Kiosk & Public Views**
   - File target: `scan.html`, `scan_sholat.html`, `attendance_face.html`, `face_register.html`, `login.html`, `idcard_single.html`.

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
  go build -o /dev/null
  ```

---

## 💡 Instruksi Khusus untuk AI Agent Saat Melakukan Edit
1. Lakukan perubahan per bagian/modul kecil secara rapi.
2. Selalu cek kembali bahwa tag penutup `{{end}}` dan Go template blocks (`{{define "..."}}`, `{{template "..." .}}`) tidak rusak.
3. Setelah mengedit tampilan, periksa apakah ada JavaScript inline yang terdampak class name barunya.
4. Jangan menghapus komentar fungsional yang sudah ada di template.

