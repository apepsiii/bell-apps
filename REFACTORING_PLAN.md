# SmartBell Refactoring Plan

## Current State (After Refactoring)

```
bell/
├── main.go                     # ~3200 lines - monolithic (uses old handler methods)
├── announcement_handlers.go    # ~160 lines (STILL NEEDED - method on App)
├── qr_handlers.go              # ~294 lines (STILL NEEDED)
├── holiday_handlers.go         # ~321 lines (STILL NEEDED)
├── operator_handlers.go        # ~265 lines (STILL NEEDED - method on App)
├── point_handlers.go          # ~357 lines (STILL NEEDED - method on App)
├── prayer_handlers.go          # (moved to internal/handler/)
├── report_handlers.go          # ~132 lines (STILL NEEDED)
├── report_helpers.go           # ~205 lines (STILL NEEDED)
├── report_pdf.go              # ~135 lines (moved to pkg/pdf/)
├── seed_operator.go            # ~41 lines (moved to internal/repository/)
├── internal/
│   ├── app/App.go             # App struct definition
│   ├── config/config.go        # Configuration constants
│   ├── models/models.go       # All data models
│   ├── handler/
│   │   ├── announcement.go    # REFACTORED - standalone functions
│   │   ├── holiday.go         # REFACTORED - standalone functions
│   │   ├── operator.go        # REFACTORED - standalone functions
│   │   ├── point.go           # REFACTORED - standalone functions
│   │   ├── prayer.go         # REFACTORED - standalone functions
│   │   └── report.go         # REFACTORED - standalone functions
│   └── repository/
│       ├── db.go              # Database initialization
│       └── seed.go            # Operator seeding
├── pkg/
│   ├── utils/date.go
│   ├── qrcode/qr.go
│   ├── onesender/client.go
│   └── pdf/report.go
└── migrations/
    └── 001_initial_schema.sql
```

---

## Target Structure

```
bell/
├── cmd/
│   └── server/
│       └── main.go             # Minimal entry point (~200 lines)
├── internal/
│   ├── app/                    # Application struct & state
│   │   └── App.go
│   ├── config/                 # Configuration constants
│   │   └── config.go
│   ├── models/                 # Data models & types
│   │   └── models.go
│   ├── repository/             # Database operations
│   │   └── db.go
│   ├── service/               # Business logic
│   │   └── service.go
│   ├── handler/               # HTTP handlers (REFACTORED)
│   │   ├── announcement.go
│   │   ├── attendance.go
│   │   ├── class.go
│   │   ├── device.go
│   │   ├── holiday.go
│   │   ├── major.go
│   │   ├── operator.go
│   │   ├── prayer.go
│   │   ├── point.go
│   │   ├── report.go
│   │   ├── schedule.go
│   │   ├── staff.go
│   │   ├── student.go
│   │   └── template.go        # Template renderer
│   ├── middleware/             # HTTP middleware
│   │   ├── auth.go
│   │   └── middleware.go
│   └── router/                # Route definitions
│       └── router.go
├── pkg/                       # Reusable packages
│   ├── utils/
│   ├── qrcode/
│   ├── onesender/
│   └── pdf/
├── migrations/                 # SQL schema
│   └── 001_initial_schema.sql
├── views/                      # HTML templates (embedded)
└── public/                     # Static assets
```

---

## Refactoring Batches

### Batch 1: ✅ COMPLETED
- [x] Extract `pkg/utils/date.go` - DateToIndo, FormatPhone
- [x] Extract `pkg/qrcode/qr.go` - QR code generation
- [x] Extract `pkg/onesender/client.go` - WhatsApp client
- [x] Extract `pkg/pdf/report.go` - PDF generation
- [x] Extract `migrations/001_initial_schema.sql` - SQL schema
- [x] Extract `internal/handler/announcement.go`
- [x] Extract `internal/handler/holiday.go`
- [x] Extract `internal/handler/prayer.go`

### Batch 2: ✅ COMPLETED - Operator & Point Handlers
- [x] `internal/handler/operator.go` - From `operator_handlers.go`
- [x] `internal/handler/point.go` - From `point_handlers.go`

### Batch 3: ✅ COMPLETED - Report Handlers
- [x] `internal/handler/report.go` - From `report_handlers.go`
- [x] Extract query functions from `report_helpers.go`
- [x] Note: `report_pdf.go` already in `pkg/pdf/report.go`

### Batch 4: CRUD Handlers (Major, Class, Student, Staff, Device, Schedule)
- [ ] `internal/handler/major.go`
- [ ] `internal/handler/class.go`
- [ ] `internal/handler/student.go`
- [ ] `internal/handler/staff.go`
- [ ] `internal/handler/device.go`
- [ ] `internal/handler/schedule.go`

### Batch 5: Attendance & Common Handlers
- [ ] `internal/handler/attendance.go` - Main attendance logic (from main.go)
- [ ] `internal/handler/template.go` - Template renderer
- [ ] `internal/handler/pages.go` - Page handlers (dashboard, scan, profile, etc.)

### Batch 6: ✅ COMPLETED - Models & Config
- [x] `internal/models/models.go` - All data models from main.go
- [x] `internal/config/config.go` - Configuration constants

### Batch 7: ✅ COMPLETED - Database & Repository
- [x] `internal/repository/db.go` - DB initialization & migrations
- [x] Move `IsWorkingDay`, `GetWorkingDaysInMonth` helpers

### Batch 8: Extract Middleware & Router
- [ ] `internal/middleware/auth.go` - Auth middleware
- [ ] `internal/router/router.go` - All route definitions

### Batch 9: Wire main.go to use refactored packages (PARTIAL)
- [x] main.go now imports `internal/repository`
- [x] main.go uses `repository.InitDB()` instead of inline sql.Open
- [ ] Wire handlers from internal/handler/ packages
- [ ] Keep embedded files (views, scripts)

### Batch 10: Final Cleanup (PENDING)
- [ ] Remove old handler files from root (only after Batch 9 complete)
- [ ] Verify compilation
- [ ] Test all endpoints
- [ ] Update documentation

## NOTE: Old handler files still needed
The following files are still in root because main.go uses them as methods on App struct:
- `operator_handlers.go` - used via `app.OperatorLoginHandler`
- `point_handlers.go` - used via `app.GetPointRulesHandler`
- `report_handlers.go` - used via `app.DailyReportHandler`
- `announcement_handlers.go` - used via `app.GetAnnouncementsHandler`
- `holiday_handlers.go` - used via `app.GetHolidaysHandler`
- `qr_handlers.go` - used via `app.GenerateQRCodeHandler`
- `prayer_handlers.go` - used via `app.GetPrayerLogsHandler`

These will be removed only after main.go is fully wired to use internal/handler/ functions.

---

## Handler Mapping

| Old File | New Location | Notes |
|----------|-------------|-------|
| `main.go` (3204 lines) | `cmd/server/main.go` | Strip to entry point only |
| `operator_handlers.go` | `internal/handler/operator.go` | Login, logout, profile |
| `point_handlers.go` | `internal/handler/point.go` | Rules, rewards, transactions |
| `report_handlers.go` | `internal/handler/report.go` | PDF report generation |
| `report_helpers.go` | `internal/service/report_service.go` | Query functions |
| `seed_operator.go` | `internal/repository/seed.go` | Default data seeding |
| CRUD in main.go | `internal/handler/{major,class,student,staff,device,schedule}.go` | Split by entity |
| Attendance in main.go | `internal/handler/attendance.go` | Main attendance logic |
| Page handlers | `internal/handler/pages.go` | Dashboard, scan, profile |

---

## Key Patterns

### Handler Pattern (After Refactor)
```go
// Before (method on App)
func (a *App) GetAnnouncementsHandler(c echo.Context) error {
    rows, err := a.DB.Query("SELECT ...")
    ...
}

// After (standalone function)
func GetAnnouncements(db *sql.DB) echo.HandlerFunc {
    return func(c echo.Context) error {
        rows, err := db.Query("SELECT ...")
        ...
    }
}
```

### Route Wiring (After Refactor)
```go
// Before (in main.go)
e.GET("/api/announcements", app.GetAnnouncementsHandler)

// After (in main.go or router.go)
e.GET("/api/announcements", handler.GetAnnouncements(app.DB))
```

---

## Verification Steps

After each batch:
1. `go build` - Verify compilation
2. `go vet` - Check for issues
3. `go test ./...` - Run tests if any
4. Manual testing of affected endpoints

---

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Breaking existing routes | Keep old files until new ones verified |
| Circular dependencies | Follow clean architecture: handler -> service -> repository |
| Too many small files | Group related handlers together |
| Losing context (db, app state) | Pass dependencies explicitly via function params or context |

---

## Timeline Estimate

- Batch 1: ✅ Done
- Batch 2-5: 2-3 hours
- Batch 6-8: 1-2 hours
- Batch 9-10: 1-2 hours
- **Total: ~5-7 hours**
