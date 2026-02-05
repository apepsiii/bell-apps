# Rencana Pengembangan: Fitur Laporan Kehadiran PDF

## Overview
Menambahkan fitur generate laporan kehadiran dalam format PDF dengan pilihan periode:
- **Harian** - Laporan kehadiran satu hari tertentu
- **Mingguan** - Laporan kehadiran dalam rentang 7 hari
- **Bulanan** - Laporan kehadiran per bulan

Fitur ini akan terintegrasi di dashboard admin dengan datepicker untuk memilih rentang waktu yang diinginkan.

---

## Timeline Estimasi
- **Total Waktu**: 8-12 jam development
- **Fase 1**: Backend & PDF Generator (4-5 jam)
- **Fase 2**: Frontend UI (2-3 jam)
- **Fase 3**: Testing & Polish (2-4 jam)

---

## Tech Stack Tambahan

### Backend
- **PDF Library**: `github.com/jung-kurt/gofpdf` (sudah ada di project)
- **Time Parsing**: Go standard library `time`
- **Query Aggregation**: SQLite dengan GROUP BY dan DATE functions

### Frontend
- **Datepicker**: Flatpickr.js (lightweight, 10KB gzipped)
  - Alternative: Native HTML5 `<input type="date">` untuk simplicity
- **Icons**: Lucide Icons (sudah ada)
- **Loading State**: SweetAlert2 (sudah ada)

---

## Database Schema (Sudah Ada)

Tidak ada perubahan schema diperlukan. Menggunakan tabel existing:

```sql
-- Table: attendance_logs
CREATE TABLE attendance_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER,
    staff_id INTEGER,
    timestamp TEXT,
    status TEXT,  -- 'datang', 'terlambat', 'pulang', 'sakit', 'izin', 'alpha'
    method TEXT,  -- 'RFID', 'MANUAL'
    FOREIGN KEY (student_id) REFERENCES students(id),
    FOREIGN KEY (staff_id) REFERENCES staff(id)
)

-- Table: students
CREATE TABLE students (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nis TEXT UNIQUE,
    name TEXT,
    class_id INTEGER,
    rfid TEXT,
    phone_parent TEXT,
    photo TEXT
)

-- Table: staff
CREATE TABLE staff (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nip TEXT UNIQUE,
    name TEXT,
    role TEXT,
    rfid TEXT,
    phone TEXT
)
```

---

## API Endpoints (Baru)

### 1. Generate Report Harian
```
GET /admin/report/daily?date=2024-02-03&type=student
```

**Query Parameters:**
- `date` (required): Tanggal dalam format YYYY-MM-DD
- `type` (required): `student` atau `staff`
- `class_id` (optional): Filter berdasarkan kelas tertentu
- `format` (optional): `pdf` (default) atau `json` untuk preview

**Response:**
- PDF file download dengan nama: `Laporan_Harian_Siswa_2024-02-03.pdf`
- Header: `Content-Type: application/pdf`

**Isi Report:**
- Header: Logo sekolah, judul, tanggal
- Statistik: Total siswa/staff, Hadir, Terlambat, Sakit, Izin, Alpha
- Tabel: No, NIS/NIP, Nama, Kelas/Role, Status, Waktu Kedatangan
- Footer: Tanggal cetak, total halaman

---

### 2. Generate Report Mingguan
```
GET /admin/report/weekly?start=2024-01-29&end=2024-02-04&type=student
```

**Query Parameters:**
- `start` (required): Tanggal mulai (YYYY-MM-DD)
- `end` (required): Tanggal akhir (YYYY-MM-DD)
- `type` (required): `student` atau `staff`
- `class_id` (optional): Filter berdasarkan kelas

**Response:**
- PDF file: `Laporan_Mingguan_Siswa_29Jan-04Feb2024.pdf`

**Isi Report:**
- Header: Logo, judul, periode
- Statistik Keseluruhan:
  - Total hari sekolah dalam periode
  - Rata-rata kehadiran per hari
  - Persentase kehadiran
- Tabel Summary per Siswa/Staff:
  - No, NIS/NIP, Nama, Hadir, Terlambat, Sakit, Izin, Alpha, %Kehadiran
- Chart: Bar chart perbandingan status (optional, bisa ditambahkan nanti)

---

### 3. Generate Report Bulanan
```
GET /admin/report/monthly?month=2024-02&type=student
```

**Query Parameters:**
- `month` (required): Bulan dalam format YYYY-MM
- `type` (required): `student` atau `staff`
- `class_id` (optional): Filter berdasarkan kelas

**Response:**
- PDF file: `Laporan_Bulanan_Siswa_Februari2024.pdf`

**Isi Report:**
- Header: Logo, judul "Laporan Kehadiran Bulan Februari 2024"
- Statistik Bulanan:
  - Total hari sekolah
  - Total siswa/staff
  - Persentase kehadiran rata-rata
  - Trend kehadiran (meningkat/menurun dari bulan lalu)
- Tabel Detail per Individu:
  - No, NIS/NIP, Nama, Total Hadir, Terlambat, Sakit, Izin, Alpha, %Kehadiran
- Keterangan: Siswa dengan kehadiran < 80% (highlight merah)
- Footer: TTD Kepala Sekolah (placeholder)

---

## Implementasi Backend

### File: `main.go`

#### 1. Import Library Tambahan
```go
import (
    "github.com/jung-kurt/gofpdf"
    "time"
    "fmt"
    "strconv"
)
```

#### 2. Struct untuk Data Report
```go
type ReportData struct {
    Title       string
    Period      string
    GeneratedAt string
    Type        string // "student" or "staff"
    
    // Statistics
    TotalRecords    int
    TotalPresent    int
    TotalLate       int
    TotalSick       int
    TotalPermission int
    TotalAbsent     int
    AttendanceRate  float64
    
    // Details
    Records []ReportRecord
}

type ReportRecord struct {
    No              int
    ID              string // NIS or NIP
    Name            string
    ClassOrRole     string
    Status          string
    Time            string
    
    // For weekly/monthly
    PresentCount    int
    LateCount       int
    SickCount       int
    PermissionCount int
    AbsentCount     int
    AttendanceRate  float64
}
```

#### 3. Handler untuk Report Harian
```go
func handleDailyReport(c echo.Context) error {
    date := c.QueryParam("date")
    reportType := c.QueryParam("type") // "student" or "staff"
    classID := c.QueryParam("class_id")
    format := c.QueryParam("format") // "pdf" or "json"
    
    // Validate date format
    _, err := time.Parse("2006-01-02", date)
    if err != nil {
        return c.JSON(400, map[string]string{"error": "Invalid date format"})
    }
    
    // Query data from database
    data := queryDailyReport(date, reportType, classID)
    
    // Return JSON preview if requested
    if format == "json" {
        return c.JSON(200, data)
    }
    
    // Generate PDF
    pdf := generateDailyPDF(data)
    
    // Set headers
    filename := fmt.Sprintf("Laporan_Harian_%s_%s.pdf", 
        reportType, date)
    c.Response().Header().Set("Content-Type", "application/pdf")
    c.Response().Header().Set("Content-Disposition", 
        fmt.Sprintf("attachment; filename=%s", filename))
    
    return pdf.Output(c.Response().Writer)
}
```

#### 4. Query Function untuk Data Harian
```go
func queryDailyReport(date, reportType, classID string) ReportData {
    data := ReportData{
        Title:       "Laporan Kehadiran Harian",
        Period:      date,
        GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
        Type:        reportType,
    }
    
    var query string
    var args []interface{}
    
    if reportType == "student" {
        query = `
            SELECT s.nis, s.name, c.name as class_name, 
                   al.status, al.timestamp
            FROM students s
            LEFT JOIN classes c ON s.class_id = c.id
            LEFT JOIN attendance_logs al ON s.id = al.student_id 
                AND DATE(al.timestamp) = ?
            WHERE 1=1
        `
        args = append(args, date)
        
        if classID != "" {
            query += " AND s.class_id = ?"
            args = append(args, classID)
        }
        
        query += " ORDER BY c.name, s.name"
    } else {
        query = `
            SELECT st.nip, st.name, st.role,
                   al.status, al.timestamp
            FROM staff st
            LEFT JOIN attendance_logs al ON st.id = al.staff_id 
                AND DATE(al.timestamp) = ?
            ORDER BY st.name
        `
        args = append(args, date)
    }
    
    rows, _ := db.Query(query, args...)
    defer rows.Close()
    
    no := 1
    for rows.Next() {
        var record ReportRecord
        var status, timestamp sql.NullString
        
        if reportType == "student" {
            rows.Scan(&record.ID, &record.Name, &record.ClassOrRole, 
                     &status, &timestamp)
        } else {
            rows.Scan(&record.ID, &record.Name, &record.ClassOrRole, 
                     &status, &timestamp)
        }
        
        record.No = no
        if status.Valid {
            record.Status = status.String
            record.Time = timestamp.String
            
            // Count statistics
            switch status.String {
            case "datang":
                data.TotalPresent++
            case "terlambat":
                data.TotalLate++
            case "sakit":
                data.TotalSick++
            case "izin":
                data.TotalPermission++
            case "alpha":
                data.TotalAbsent++
            }
        } else {
            record.Status = "Tidak Hadir"
            data.TotalAbsent++
        }
        
        data.Records = append(data.Records, record)
        no++
    }
    
    data.TotalRecords = len(data.Records)
    if data.TotalRecords > 0 {
        data.AttendanceRate = float64(data.TotalPresent+data.TotalLate) / 
                             float64(data.TotalRecords) * 100
    }
    
    return data
}
```

#### 5. PDF Generator Function
```go
func generateDailyPDF(data ReportData) *gofpdf.Fpdf {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    
    // Header
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(0, 10, "LAPORAN KEHADIRAN HARIAN")
    pdf.Ln(8)
    
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(0, 6, "Tanggal: "+data.Period)
    pdf.Ln(6)
    pdf.Cell(0, 6, "Tipe: "+strings.Title(data.Type))
    pdf.Ln(10)
    
    // Statistics Box
    pdf.SetFont("Arial", "B", 10)
    pdf.SetFillColor(240, 240, 240)
    pdf.CellFormat(40, 7, "Total", "1", 0, "C", true, 0, "")
    pdf.CellFormat(30, 7, "Hadir", "1", 0, "C", true, 0, "")
    pdf.CellFormat(30, 7, "Terlambat", "1", 0, "C", true, 0, "")
    pdf.CellFormat(30, 7, "Sakit", "1", 0, "C", true, 0, "")
    pdf.CellFormat(30, 7, "Izin", "1", 0, "C", true, 0, "")
    pdf.CellFormat(30, 7, "Alpha", "1", 1, "C", true, 0, "")
    
    pdf.SetFont("Arial", "", 10)
    pdf.Cell(40, 7, strconv.Itoa(data.TotalRecords), "1", 0, "C", false, 0, "")
    pdf.Cell(30, 7, strconv.Itoa(data.TotalPresent), "1", 0, "C", false, 0, "")
    pdf.Cell(30, 7, strconv.Itoa(data.TotalLate), "1", 0, "C", false, 0, "")
    pdf.Cell(30, 7, strconv.Itoa(data.TotalSick), "1", 0, "C", false, 0, "")
    pdf.Cell(30, 7, strconv.Itoa(data.TotalPermission), "1", 0, "C", false, 0, "")
    pdf.Cell(30, 7, strconv.Itoa(data.TotalAbsent), "1", 1, "C", false, 0, "")
    pdf.Ln(10)
    
    // Table Header
    pdf.SetFont("Arial", "B", 9)
    pdf.SetFillColor(200, 220, 255)
    pdf.CellFormat(10, 7, "No", "1", 0, "C", true, 0, "")
    pdf.CellFormat(30, 7, "NIS/NIP", "1", 0, "C", true, 0, "")
    pdf.CellFormat(60, 7, "Nama", "1", 0, "C", true, 0, "")
    pdf.CellFormat(40, 7, "Kelas/Role", "1", 0, "C", true, 0, "")
    pdf.CellFormat(25, 7, "Status", "1", 0, "C", true, 0, "")
    pdf.CellFormat(25, 7, "Waktu", "1", 1, "C", true, 0, "")
    
    // Table Data
    pdf.SetFont("Arial", "", 8)
    for _, record := range data.Records {
        pdf.Cell(10, 6, strconv.Itoa(record.No), "1", 0, "C", false, 0, "")
        pdf.Cell(30, 6, record.ID, "1", 0, "L", false, 0, "")
        pdf.Cell(60, 6, record.Name, "1", 0, "L", false, 0, "")
        pdf.Cell(40, 6, record.ClassOrRole, "1", 0, "L", false, 0, "")
        pdf.Cell(25, 6, record.Status, "1", 0, "C", false, 0, "")
        
        timeStr := ""
        if record.Time != "" {
            t, _ := time.Parse("2006-01-02 15:04:05", record.Time)
            timeStr = t.Format("15:04")
        }
        pdf.Cell(25, 6, timeStr, "1", 1, "C", false, 0, "")
    }
    
    // Footer
    pdf.Ln(10)
    pdf.SetFont("Arial", "I", 8)
    pdf.Cell(0, 6, "Dicetak pada: "+data.GeneratedAt)
    
    return pdf
}
```

#### 6. Handler untuk Report Mingguan
```go
func handleWeeklyReport(c echo.Context) error {
    startDate := c.QueryParam("start")
    endDate := c.QueryParam("end")
    reportType := c.QueryParam("type")
    classID := c.QueryParam("class_id")
    
    // Validate dates
    _, err1 := time.Parse("2006-01-02", startDate)
    _, err2 := time.Parse("2006-01-02", endDate)
    if err1 != nil || err2 != nil {
        return c.JSON(400, map[string]string{"error": "Invalid date format"})
    }
    
    // Query aggregated data
    data := queryWeeklyReport(startDate, endDate, reportType, classID)
    
    // Generate PDF
    pdf := generateWeeklyPDF(data)
    
    filename := fmt.Sprintf("Laporan_Mingguan_%s_%s_to_%s.pdf", 
        reportType, startDate, endDate)
    c.Response().Header().Set("Content-Type", "application/pdf")
    c.Response().Header().Set("Content-Disposition", 
        fmt.Sprintf("attachment; filename=%s", filename))
    
    return pdf.Output(c.Response().Writer)
}
```

#### 7. Query Function untuk Data Mingguan
```go
func queryWeeklyReport(startDate, endDate, reportType, classID string) ReportData {
    data := ReportData{
        Title:       "Laporan Kehadiran Mingguan",
        Period:      startDate + " s/d " + endDate,
        GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
        Type:        reportType,
    }
    
    var query string
    var args []interface{}
    
    if reportType == "student" {
        query = `
            SELECT s.nis, s.name, c.name as class_name,
                SUM(CASE WHEN al.status = 'datang' THEN 1 ELSE 0 END) as present,
                SUM(CASE WHEN al.status = 'terlambat' THEN 1 ELSE 0 END) as late,
                SUM(CASE WHEN al.status = 'sakit' THEN 1 ELSE 0 END) as sick,
                SUM(CASE WHEN al.status = 'izin' THEN 1 ELSE 0 END) as permission,
                SUM(CASE WHEN al.status = 'alpha' THEN 1 ELSE 0 END) as absent
            FROM students s
            LEFT JOIN classes c ON s.class_id = c.id
            LEFT JOIN attendance_logs al ON s.id = al.student_id 
                AND DATE(al.timestamp) BETWEEN ? AND ?
            WHERE 1=1
        `
        args = append(args, startDate, endDate)
        
        if classID != "" {
            query += " AND s.class_id = ?"
            args = append(args, classID)
        }
        
        query += " GROUP BY s.id ORDER BY c.name, s.name"
    } else {
        query = `
            SELECT st.nip, st.name, st.role,
                SUM(CASE WHEN al.status = 'datang' THEN 1 ELSE 0 END) as present,
                SUM(CASE WHEN al.status = 'terlambat' THEN 1 ELSE 0 END) as late,
                SUM(CASE WHEN al.status = 'sakit' THEN 1 ELSE 0 END) as sick,
                SUM(CASE WHEN al.status = 'izin' THEN 1 ELSE 0 END) as permission,
                SUM(CASE WHEN al.status = 'alpha' THEN 1 ELSE 0 END) as absent
            FROM staff st
            LEFT JOIN attendance_logs al ON st.id = al.staff_id 
                AND DATE(al.timestamp) BETWEEN ? AND ?
            GROUP BY st.id ORDER BY st.name
        `
        args = append(args, startDate, endDate)
    }
    
    rows, _ := db.Query(query, args...)
    defer rows.Close()
    
    no := 1
    totalDays := calculateSchoolDays(startDate, endDate) // Custom function
    
    for rows.Next() {
        var record ReportRecord
        
        rows.Scan(&record.ID, &record.Name, &record.ClassOrRole,
                 &record.PresentCount, &record.LateCount, 
                 &record.SickCount, &record.PermissionCount, 
                 &record.AbsentCount)
        
        record.No = no
        
        // Calculate attendance rate
        totalAttendance := record.PresentCount + record.LateCount
        if totalDays > 0 {
            record.AttendanceRate = float64(totalAttendance) / 
                                   float64(totalDays) * 100
        }
        
        data.Records = append(data.Records, record)
        
        // Aggregate statistics
        data.TotalPresent += record.PresentCount
        data.TotalLate += record.LateCount
        data.TotalSick += record.SickCount
        data.TotalPermission += record.PermissionCount
        data.TotalAbsent += record.AbsentCount
        
        no++
    }
    
    data.TotalRecords = len(data.Records)
    
    return data
}

func calculateSchoolDays(startDate, endDate string) int {
    start, _ := time.Parse("2006-01-02", startDate)
    end, _ := time.Parse("2006-01-02", endDate)
    
    days := 0
    for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
        // Exclude Saturday (6) and Sunday (0)
        if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
            days++
        }
    }
    return days
}
```

#### 8. PDF Generator untuk Report Mingguan/Bulanan
```go
func generateWeeklyPDF(data ReportData) *gofpdf.Fpdf {
    pdf := gofpdf.New("L", "mm", "A4", "") // Landscape for more columns
    pdf.AddPage()
    
    // Header
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(0, 10, strings.ToUpper(data.Title))
    pdf.Ln(8)
    
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(0, 6, "Periode: "+data.Period)
    pdf.Ln(6)
    pdf.Cell(0, 6, "Tipe: "+strings.Title(data.Type))
    pdf.Ln(10)
    
    // Table Header
    pdf.SetFont("Arial", "B", 9)
    pdf.SetFillColor(200, 220, 255)
    pdf.CellFormat(10, 7, "No", "1", 0, "C", true, 0, "")
    pdf.CellFormat(30, 7, "NIS/NIP", "1", 0, "C", true, 0, "")
    pdf.CellFormat(60, 7, "Nama", "1", 0, "C", true, 0, "")
    pdf.CellFormat(35, 7, "Kelas/Role", "1", 0, "C", true, 0, "")
    pdf.CellFormat(20, 7, "Hadir", "1", 0, "C", true, 0, "")
    pdf.CellFormat(25, 7, "Terlambat", "1", 0, "C", true, 0, "")
    pdf.CellFormat(20, 7, "Sakit", "1", 0, "C", true, 0, "")
    pdf.CellFormat(20, 7, "Izin", "1", 0, "C", true, 0, "")
    pdf.CellFormat(20, 7, "Alpha", "1", 0, "C", true, 0, "")
    pdf.CellFormat(30, 7, "% Kehadiran", "1", 1, "C", true, 0, "")
    
    // Table Data
    pdf.SetFont("Arial", "", 8)
    for _, record := range data.Records {
        pdf.Cell(10, 6, strconv.Itoa(record.No), "1", 0, "C", false, 0, "")
        pdf.Cell(30, 6, record.ID, "1", 0, "L", false, 0, "")
        pdf.Cell(60, 6, record.Name, "1", 0, "L", false, 0, "")
        pdf.Cell(35, 6, record.ClassOrRole, "1", 0, "L", false, 0, "")
        pdf.Cell(20, 6, strconv.Itoa(record.PresentCount), "1", 0, "C", false, 0, "")
        pdf.Cell(25, 6, strconv.Itoa(record.LateCount), "1", 0, "C", false, 0, "")
        pdf.Cell(20, 6, strconv.Itoa(record.SickCount), "1", 0, "C", false, 0, "")
        pdf.Cell(20, 6, strconv.Itoa(record.PermissionCount), "1", 0, "C", false, 0, "")
        pdf.Cell(20, 6, strconv.Itoa(record.AbsentCount), "1", 0, "C", false, 0, "")
        
        // Color code attendance rate
        if record.AttendanceRate < 80 {
            pdf.SetTextColor(255, 0, 0) // Red for low attendance
        }
        pdf.Cell(30, 6, fmt.Sprintf("%.1f%%", record.AttendanceRate), "1", 1, "C", false, 0, "")
        pdf.SetTextColor(0, 0, 0) // Reset to black
    }
    
    // Footer
    pdf.Ln(10)
    pdf.SetFont("Arial", "I", 8)
    pdf.Cell(0, 6, "Dicetak pada: "+data.GeneratedAt)
    
    return pdf
}
```

---

## Implementasi Frontend

### File: `views/admin.html`

#### 1. Tambahkan Menu "Laporan" di Sidebar
```html
<!-- Setelah menu Attendance -->
<a href="#" onclick="showSection('report')" 
   class="flex items-center px-4 py-3 text-gray-300 hover:bg-blue-700">
    <i data-lucide="file-text" class="w-5 h-5 mr-3"></i>
    Laporan
</a>
```

#### 2. Tambahkan Section Report Content
```html
<!-- Report Section -->
<div id="reportSection" class="section hidden">
    <div class="mb-6">
        <h2 class="text-2xl font-bold text-gray-800">Laporan Kehadiran</h2>
        <p class="text-gray-600">Generate laporan kehadiran dalam format PDF</p>
    </div>

    <!-- Filter Form -->
    <div class="bg-white rounded-lg shadow-md p-6 mb-6">
        <h3 class="text-lg font-semibold mb-4">Filter Laporan</h3>
        
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <!-- Report Type -->
            <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                    Tipe Laporan
                </label>
                <select id="reportPeriodType" 
                        class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                        onchange="toggleDateInputs()">
                    <option value="daily">Harian</option>
                    <option value="weekly">Mingguan</option>
                    <option value="monthly">Bulanan</option>
                </select>
            </div>

            <!-- Subject Type -->
            <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                    Subjek
                </label>
                <select id="reportSubjectType" 
                        class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500">
                    <option value="student">Siswa</option>
                    <option value="staff">Staff</option>
                </select>
            </div>

            <!-- Class Filter (for students only) -->
            <div id="classFilterDiv">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                    Kelas (Opsional)
                </label>
                <select id="reportClassFilter" 
                        class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500">
                    <option value="">Semua Kelas</option>
                    <!-- Populated dynamically -->
                </select>
            </div>
        </div>

        <!-- Date Inputs -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
            <!-- Daily Date -->
            <div id="dailyDateDiv">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                    Tanggal
                </label>
                <input type="date" id="reportDailyDate" 
                       class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                       value="<?php echo date('Y-m-d'); ?>">
            </div>

            <!-- Weekly Start Date -->
            <div id="weeklyStartDiv" class="hidden">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                    Tanggal Mulai
                </label>
                <input type="date" id="reportWeeklyStart" 
                       class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500">
            </div>

            <!-- Weekly End Date -->
            <div id="weeklyEndDiv" class="hidden">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                    Tanggal Akhir
                </label>
                <input type="date" id="reportWeeklyEnd" 
                       class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500">
            </div>

            <!-- Monthly Month -->
            <div id="monthlyMonthDiv" class="hidden">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                    Bulan
                </label>
                <input type="month" id="reportMonthlyMonth" 
                       class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                       value="<?php echo date('Y-m'); ?>">
            </div>
        </div>

        <!-- Action Buttons -->
        <div class="flex gap-3 mt-6">
            <button onclick="generateReport('preview')" 
                    class="px-6 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 flex items-center gap-2">
                <i data-lucide="eye" class="w-4 h-4"></i>
                Preview
            </button>
            <button onclick="generateReport('download')" 
                    class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2">
                <i data-lucide="download" class="w-4 h-4"></i>
                Download PDF
            </button>
        </div>
    </div>

    <!-- Preview Area -->
    <div id="reportPreview" class="bg-white rounded-lg shadow-md p-6 hidden">
        <div class="flex justify-between items-center mb-4">
            <h3 class="text-lg font-semibold">Preview Laporan</h3>
            <button onclick="closePreview()" class="text-gray-500 hover:text-gray-700">
                <i data-lucide="x" class="w-5 h-5"></i>
            </button>
        </div>
        <div id="reportPreviewContent" class="overflow-x-auto">
            <!-- Preview content will be loaded here -->
        </div>
    </div>
</div>
```

#### 3. Tambahkan JavaScript untuk Report
```html
<script>
// Toggle date inputs based on report type
function toggleDateInputs() {
    const periodType = document.getElementById('reportPeriodType').value;
    
    // Hide all
    document.getElementById('dailyDateDiv').classList.add('hidden');
    document.getElementById('weeklyStartDiv').classList.add('hidden');
    document.getElementById('weeklyEndDiv').classList.add('hidden');
    document.getElementById('monthlyMonthDiv').classList.add('hidden');
    
    // Show relevant
    if (periodType === 'daily') {
        document.getElementById('dailyDateDiv').classList.remove('hidden');
    } else if (periodType === 'weekly') {
        document.getElementById('weeklyStartDiv').classList.remove('hidden');
        document.getElementById('weeklyEndDiv').classList.remove('hidden');
    } else if (periodType === 'monthly') {
        document.getElementById('monthlyMonthDiv').classList.remove('hidden');
    }
    
    lucide.createIcons(); // Refresh icons
}

// Toggle class filter based on subject type
document.getElementById('reportSubjectType').addEventListener('change', function() {
    const classFilterDiv = document.getElementById('classFilterDiv');
    if (this.value === 'student') {
        classFilterDiv.classList.remove('hidden');
    } else {
        classFilterDiv.classList.add('hidden');
    }
});

// Populate class filter
async function loadClassesForReport() {
    const response = await fetch('/admin/classes');
    const classes = await response.json();
    
    const select = document.getElementById('reportClassFilter');
    select.innerHTML = '<option value="">Semua Kelas</option>';
    
    classes.forEach(cls => {
        const option = document.createElement('option');
        option.value = cls.id;
        option.textContent = cls.name;
        select.appendChild(option);
    });
}

// Generate Report
async function generateReport(action) {
    const periodType = document.getElementById('reportPeriodType').value;
    const subjectType = document.getElementById('reportSubjectType').value;
    const classID = document.getElementById('reportClassFilter').value;
    
    let url = '';
    let params = new URLSearchParams({
        type: subjectType
    });
    
    if (classID) params.append('class_id', classID);
    if (action === 'preview') params.append('format', 'json');
    
    if (periodType === 'daily') {
        const date = document.getElementById('reportDailyDate').value;
        if (!date) {
            Swal.fire('Error', 'Pilih tanggal terlebih dahulu', 'error');
            return;
        }
        params.append('date', date);
        url = '/admin/report/daily?' + params.toString();
    } else if (periodType === 'weekly') {
        const start = document.getElementById('reportWeeklyStart').value;
        const end = document.getElementById('reportWeeklyEnd').value;
        if (!start || !end) {
            Swal.fire('Error', 'Pilih tanggal mulai dan akhir', 'error');
            return;
        }
        params.append('start', start);
        params.append('end', end);
        url = '/admin/report/weekly?' + params.toString();
    } else if (periodType === 'monthly') {
        const month = document.getElementById('reportMonthlyMonth').value;
        if (!month) {
            Swal.fire('Error', 'Pilih bulan terlebih dahulu', 'error');
            return;
        }
        params.append('month', month);
        url = '/admin/report/monthly?' + params.toString();
    }
    
    if (action === 'preview') {
        // Preview mode
        Swal.fire({
            title: 'Loading...',
            text: 'Memuat preview laporan',
            allowOutsideClick: false,
            didOpen: () => Swal.showLoading()
        });
        
        const response = await fetch(url);
        const data = await response.json();
        
        Swal.close();
        showPreview(data);
    } else {
        // Download mode
        Swal.fire({
            title: 'Generating PDF...',
            text: 'Mohon tunggu',
            allowOutsideClick: false,
            didOpen: () => Swal.showLoading()
        });
        
        // Download PDF
        window.location.href = url;
        
        setTimeout(() => {
            Swal.close();
            Swal.fire('Success', 'Laporan berhasil diunduh', 'success');
        }, 2000);
    }
}

// Show preview
function showPreview(data) {
    const previewDiv = document.getElementById('reportPreview');
    const contentDiv = document.getElementById('reportPreviewContent');
    
    let html = `
        <div class="mb-4">
            <h4 class="text-xl font-bold">${data.Title}</h4>
            <p class="text-gray-600">Periode: ${data.Period}</p>
        </div>
        
        <!-- Statistics -->
        <div class="grid grid-cols-3 md:grid-cols-6 gap-4 mb-6">
            <div class="bg-blue-100 p-3 rounded text-center">
                <div class="text-2xl font-bold text-blue-600">${data.TotalPresent}</div>
                <div class="text-xs text-gray-600">Hadir</div>
            </div>
            <div class="bg-yellow-100 p-3 rounded text-center">
                <div class="text-2xl font-bold text-yellow-600">${data.TotalLate}</div>
                <div class="text-xs text-gray-600">Terlambat</div>
            </div>
            <div class="bg-green-100 p-3 rounded text-center">
                <div class="text-2xl font-bold text-green-600">${data.TotalSick}</div>
                <div class="text-xs text-gray-600">Sakit</div>
            </div>
            <div class="bg-purple-100 p-3 rounded text-center">
                <div class="text-2xl font-bold text-purple-600">${data.TotalPermission}</div>
                <div class="text-xs text-gray-600">Izin</div>
            </div>
            <div class="bg-red-100 p-3 rounded text-center">
                <div class="text-2xl font-bold text-red-600">${data.TotalAbsent}</div>
                <div class="text-xs text-gray-600">Alpha</div>
            </div>
            <div class="bg-gray-100 p-3 rounded text-center">
                <div class="text-2xl font-bold text-gray-600">${data.AttendanceRate.toFixed(1)}%</div>
                <div class="text-xs text-gray-600">Kehadiran</div>
            </div>
        </div>
        
        <!-- Table -->
        <table class="w-full border-collapse border">
            <thead class="bg-gray-100">
                <tr>
                    <th class="border px-2 py-2">No</th>
                    <th class="border px-2 py-2">NIS/NIP</th>
                    <th class="border px-2 py-2">Nama</th>
                    <th class="border px-2 py-2">Kelas/Role</th>
    `;
    
    if (data.Records[0] && data.Records[0].Status) {
        // Daily report
        html += `
                    <th class="border px-2 py-2">Status</th>
                    <th class="border px-2 py-2">Waktu</th>
        `;
    } else {
        // Weekly/Monthly report
        html += `
                    <th class="border px-2 py-2">Hadir</th>
                    <th class="border px-2 py-2">Terlambat</th>
                    <th class="border px-2 py-2">Sakit</th>
                    <th class="border px-2 py-2">Izin</th>
                    <th class="border px-2 py-2">Alpha</th>
                    <th class="border px-2 py-2">% Hadir</th>
        `;
    }
    
    html += `
                </tr>
            </thead>
            <tbody>
    `;
    
    data.Records.forEach(record => {
        html += `<tr>
            <td class="border px-2 py-1 text-center">${record.No}</td>
            <td class="border px-2 py-1">${record.ID}</td>
            <td class="border px-2 py-1">${record.Name}</td>
            <td class="border px-2 py-1">${record.ClassOrRole}</td>
        `;
        
        if (record.Status) {
            html += `
                <td class="border px-2 py-1 text-center">${record.Status}</td>
                <td class="border px-2 py-1 text-center">${record.Time || '-'}</td>
            `;
        } else {
            const attendanceClass = record.AttendanceRate < 80 ? 'text-red-600 font-bold' : '';
            html += `
                <td class="border px-2 py-1 text-center">${record.PresentCount}</td>
                <td class="border px-2 py-1 text-center">${record.LateCount}</td>
                <td class="border px-2 py-1 text-center">${record.SickCount}</td>
                <td class="border px-2 py-1 text-center">${record.PermissionCount}</td>
                <td class="border px-2 py-1 text-center">${record.AbsentCount}</td>
                <td class="border px-2 py-1 text-center ${attendanceClass}">${record.AttendanceRate.toFixed(1)}%</td>
            `;
        }
        
        html += `</tr>`;
    });
    
    html += `
            </tbody>
        </table>
    `;
    
    contentDiv.innerHTML = html;
    previewDiv.classList.remove('hidden');
}

function closePreview() {
    document.getElementById('reportPreview').classList.add('hidden');
}

// Initialize when section loads
loadClassesForReport();
</script>
```

---

## Testing Checklist

### Backend Testing
- [ ] Test API endpoint `/admin/report/daily` dengan berbagai tanggal
- [ ] Test API endpoint `/admin/report/weekly` dengan rentang 7 hari
- [ ] Test API endpoint `/admin/report/monthly` dengan bulan berbeda
- [ ] Test filter by class_id
- [ ] Test dengan data kosong (tidak ada kehadiran)
- [ ] Test dengan data besar (1000+ siswa)
- [ ] Verify PDF generation tidak error
- [ ] Verify perhitungan statistik akurat

### Frontend Testing
- [ ] Test toggle between daily/weekly/monthly
- [ ] Test datepicker functionality
- [ ] Test class filter (show/hide based on subject type)
- [ ] Test preview mode
- [ ] Test download PDF
- [ ] Test loading states
- [ ] Test error handling (invalid dates)
- [ ] Responsive design di mobile

### Edge Cases
- [ ] Weekend dates (Saturday/Sunday)
- [ ] Bulan Februari (28/29 hari)
- [ ] Siswa tanpa kelas
- [ ] Staff tanpa kehadiran sama sekali
- [ ] Nama panjang (overflow handling)
- [ ] Special characters dalam nama

---

## Deployment Steps

1. **Update go.mod** (jika perlu)
   ```bash
   go get github.com/jung-kurt/gofpdf
   go mod tidy
   ```

2. **Test locally**
   ```bash
   go run main.go
   # Akses http://localhost:8080/admin
   ```

3. **Build for production**
   ```bash
   GOOS=linux GOARCH=amd64 go build -o bell_linux main.go
   ```

4. **Upload to VPS**
   ```bash
   scp bell_linux user@vps:/path/to/app/
   scp -r views/ user@vps:/path/to/app/
   ```

5. **Restart service**
   ```bash
   sudo systemctl restart bell
   ```

---

## Future Enhancements (Optional)

### Phase 2 Features
- [ ] Export to Excel (CSV)
- [ ] Email laporan otomatis ke kepala sekolah
- [ ] Chart visualisasi dalam PDF
- [ ] Filter by jurusan
- [ ] Perbandingan antar periode
- [ ] Analisis trend kehadiran
- [ ] Ranking kelas dengan kehadiran terbaik

### Performance Optimization
- [ ] Caching untuk report yang sering diakses
- [ ] Background job untuk report besar
- [ ] Progress bar untuk PDF generation
- [ ] Pagination untuk preview (jika data > 100)

---

## Estimasi Resource

### Development Time
- Backend (main.go): 4-5 jam
- Frontend (admin.html): 2-3 jam
- Testing & Bug fixes: 2-4 jam
- **Total**: 8-12 jam

### File Changes
- `main.go`: +500-700 baris
- `views/admin.html`: +300-400 baris
- `go.mod`: +1 dependency (jika belum ada gofpdf)

### Performance Impact
- PDF generation untuk 100 siswa: ~1-2 detik
- PDF generation untuk 1000 siswa: ~5-10 detik
- Database query: <500ms (dengan index)

---

## Notes
- Pastikan timezone server set ke Asia/Jakarta untuk konsistensi waktu
- Gunakan font Arial di PDF untuk dukungan karakter Indonesia
- Pertimbangkan menambahkan logo sekolah di header PDF
- Untuk laporan bulanan, tambahkan fitur "Keterangan" untuk siswa dengan kehadiran rendah
- Consider adding watermark "CONFIDENTIAL" untuk laporan staff

---

**Status**: Ready for Implementation
**Priority**: High
**Assignee**: TBD
**Start Date**: TBD
**Target Completion**: TBD
