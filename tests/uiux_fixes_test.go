package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- UIUX-01: Pinch-to-Zoom Accessibility on Mobile (WCAG 1.4.4) ---

func TestUIUX01_Viewport_Zoom_Enabled(t *testing.T) {
	// 1. Check views/student/app.html
	appBytes, err := os.ReadFile("views/student/app.html")
	assert.NoError(t, err)
	appContent := string(appBytes)

	assert.NotContains(t, appContent, "user-scalable=no",
		"views/student/app.html must NOT disable user scaling (WCAG 1.4.4)")
	assert.NotContains(t, appContent, "maximum-scale=1.0",
		"views/student/app.html must NOT restrict maximum-scale to 1.0 (WCAG 1.4.4)")
	assert.Contains(t, appContent, `<meta name="viewport" content="width=device-width, initial-scale=1.0">`,
		"views/student/app.html must use standard responsive viewport meta tag")

	// 2. Check views/student/login.html
	loginBytes, err := os.ReadFile("views/student/login.html")
	assert.NoError(t, err)
	loginContent := string(loginBytes)

	assert.NotContains(t, loginContent, "user-scalable=no",
		"views/student/login.html must NOT disable user scaling (WCAG 1.4.4)")
	assert.NotContains(t, loginContent, "maximum-scale=1.0",
		"views/student/login.html must NOT restrict maximum-scale to 1.0 (WCAG 1.4.4)")
	assert.Contains(t, loginContent, `<meta name="viewport" content="width=device-width, initial-scale=1.0">`,
		"views/student/login.html must use standard responsive viewport meta tag")
}

// --- UIUX-02: Signage TV Typo & Text Contrast ---

func TestUIUX02_Signage_Typo_And_Contrast(t *testing.T) {
	indexBytes, err := os.ReadFile("views/index.html")
	assert.NoError(t, err)
	indexContent := string(indexBytes)

	// Check typo resolution
	assert.NotContains(t, indexContent, "AKTIFKAN SISTEN",
		"views/index.html should not have typo 'AKTIFKAN SISTEN'")
	assert.Contains(t, indexContent, "AKTIFKAN SISTEM",
		"views/index.html should have corrected label 'AKTIFKAN SISTEM'")

	// Check contrast enhancement
	assert.NotContains(t, indexContent, "text-black/40",
		"views/index.html should not use low contrast text-black/40 on lime background")
	assert.True(t, strings.Contains(indexContent, "text-black font-bold") || strings.Contains(indexContent, "text-black/80"),
		"views/index.html should use high contrast text for leaderboard loader")
}

// --- UIUX-07: Scan Terminal Recent Logs Dynamic Background Color ---

func TestUIUX07_Scan_Card_Background(t *testing.T) {
	scanBytes, err := os.ReadFile("views/scan.html")
	assert.NoError(t, err)
	scanContent := string(scanBytes)

	// Verify broken uncompiled Tailwind arbitrary class is removed
	assert.NotContains(t, scanContent, `bg-[${bg}]`,
		"views/scan.html must NOT use uncompiled dynamic template string bg-[${bg}]")

	// Verify semantic class or direct style assignment is used
	assert.True(t, strings.Contains(scanContent, "card.style.backgroundColor = bg;") || strings.Contains(scanContent, "bg-white dark:bg-zinc-900"),
		"views/scan.html must assign valid semantic classes or direct style for cards")
}

// --- UIUX: Stisla Dashboard Redesign Verification ---

func TestUIUX_Stisla_Dashboard_Redesign(t *testing.T) {
	adminBytes, err := os.ReadFile("views/admin.html")
	assert.NoError(t, err)
	adminContent := string(adminBytes)

	// 1. Verify Stisla Visual Tokens & Color Palette
	assert.Contains(t, adminContent, "#6777ef", "views/admin.html must use Stisla signature primary color #6777ef")
	assert.Contains(t, adminContent, "#5563db", "views/admin.html must use Stisla hover color #5563db")
	assert.Contains(t, adminContent, "#f4f6f9", "views/admin.html must use Stisla background canvas #f4f6f9")
	assert.Contains(t, adminContent, "navbar-bg", "views/admin.html must include Stisla .navbar-bg banner")

	// 2. Verify Stisla Sidebar & Navigation Architecture
	assert.Contains(t, adminContent, `id="main-sidebar"`, "views/admin.html must have #main-sidebar")
	assert.Contains(t, adminContent, `id="sidebar-backdrop"`, "views/admin.html must have #sidebar-backdrop for mobile drawer")
	assert.Contains(t, adminContent, "sidebar-menu-header", "views/admin.html must include Stisla categorized menu headers")
	assert.Contains(t, adminContent, "toggleMobileSidebar()", "views/admin.html must implement toggleMobileSidebar()")

	// 3. Verify Stisla Topbar & Command Features
	assert.Contains(t, adminContent, `id="global-search-input"`, "views/admin.html must include global search input")
	assert.Contains(t, adminContent, `Ctrl+K`, "views/admin.html must feature Ctrl+K command shortcut placeholder")
	assert.Contains(t, adminContent, `id="server-clock"`, "views/admin.html must display server clock")

	// 4. Verify Command Center & Card Statistics (.card-statistic-1)
	assert.Contains(t, adminContent, "card-stat", "views/admin.html must include .card-stat icon-box cards")
	assert.Regexp(t, `Siswa\s+Hadir\s+Hari\s+Ini`, adminContent, "views/admin.html must monitor today's present students")
	assert.Regexp(t, `Belum\s+Hadir\s+/\s+Alpha`, adminContent, "views/admin.html must monitor absent students")
	assert.Regexp(t, `Bel\s+Berikutnya`, adminContent, "views/admin.html must display next bell schedule")
	assert.Contains(t, adminContent, `id="next-bell-countdown"`, "views/admin.html must feature dynamic next bell countdown")

	// 5. Verify Quick Actions Toolbar
	assert.Regexp(t, `Aksi\s+Cepat\s+Admin`, adminContent, "views/admin.html must have Quick Actions Toolbar")
	assert.Contains(t, adminContent, "openModal('modal-manual-attendance')", "views/admin.html must have quick manual attendance button")
	assert.Contains(t, adminContent, "openModal('modal-announcement')", "views/admin.html must have quick announcement button")
	assert.Contains(t, adminContent, "quickTestBell(this)", "views/admin.html must have quick bell test button")

	// 6. Verify Hardware & IoT Health Status Bar
	assert.Regexp(t, `RFID\s+Gate\s+Scanner`, adminContent, "views/admin.html must monitor gate RFID scanner")
	assert.Contains(t, adminContent, "WhatsApp OneSender", "views/admin.html must monitor WhatsApp gateway")
	assert.Regexp(t, `Mesin\s+Audio\s+Bel`, adminContent, "views/admin.html must monitor audio bell engine")

	// 7. Verify Critical Sections & Template Variables (Zero Regression)
	assert.Contains(t, adminContent, `<section id="dashboard"`, "Section #dashboard must exist")
	assert.Contains(t, adminContent, `<section id="schedules"`, "Section #schedules must exist")
	assert.Contains(t, adminContent, `<section id="attendance"`, "Section #attendance must exist")
	assert.Contains(t, adminContent, `<section id="points"`, "Section #points must exist")
	assert.Contains(t, adminContent, "{{.Stats.TotalStudents}}", "Must preserve TotalStudents template variable")
	assert.Contains(t, adminContent, "{{range .Leaderboard}}", "Must preserve Leaderboard template loop")
}

// --- UIUX-12: Admin Login Password Toggle ---

func TestUIUX12_Login_Password_Toggle(t *testing.T) {
	loginBytes, err := os.ReadFile("views/login.html")
	assert.NoError(t, err)
	loginContent := string(loginBytes)

	// Verify input id and toggle button exist
	assert.Contains(t, loginContent, `id="admin-password"`,
		"views/login.html password input must have id 'admin-password'")
	assert.Contains(t, loginContent, "toggleAdminPassword()",
		"views/login.html must contain toggleAdminPassword() button handler")
	assert.Contains(t, loginContent, "function toggleAdminPassword()",
		"views/login.html must define toggleAdminPassword() script function")
}

