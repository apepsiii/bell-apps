package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type ActivityItem struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	TimeAgo     string `json:"time_ago"`
	Timestamp   string `json:"timestamp"`
}

// GetRecentActivity returns recent system activities
func GetRecentActivity(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var activities []ActivityItem

		// Get recent attendance (last 10)
		rows, _ := db.Query(`
			SELECT 'attendance' as type, user_name, status, timestamp 
			FROM attendance_logs 
			WHERE user_type = 'Siswa'
			ORDER BY timestamp DESC 
			LIMIT 10
		`)
		for rows.Next() {
			var userName, status, timestamp string
			rows.Scan(&userName, &status, &timestamp)
			activities = append(activities, ActivityItem{
				Type:        "attendance",
				Description: fmt.Sprintf("%s - %s", userName, status),
				Timestamp:   timestamp,
			})
		}
		rows.Close()

		// Get recent point transactions (last 10)
		rows, _ = db.Query(`
			SELECT sp.points_change, sp.description, sp.timestamp, s.name
			FROM student_points sp
			JOIN students s ON sp.student_id = s.id
			ORDER BY sp.timestamp DESC
			LIMIT 10
		`)
		for rows.Next() {
			var points int
			var desc, timestamp, name string
			rows.Scan(&points, &desc, &timestamp, &name)
			sign := ""
			if points > 0 {
				sign = "+"
			}
			activities = append(activities, ActivityItem{
				Type:        "point",
				Description: fmt.Sprintf("%s: %s%d poin - %s", name, sign, points, desc),
				Timestamp:   timestamp,
			})
		}
		rows.Close()

		// Get recent English submissions (last 5 approved)
		rows, _ = db.Query(`
			SELECT es.status, es.reviewed_at, s.name, eq.title
			FROM english_submissions es
			JOIN students s ON es.student_id = s.id
			JOIN english_quests eq ON es.quest_id = eq.id
			WHERE es.status IN ('approved', 'rejected')
			ORDER BY es.reviewed_at DESC
			LIMIT 5
		`)
		for rows.Next() {
			var status, reviewedAt, name, title string
			rows.Scan(&status, &reviewedAt, &name, &title)
			statusText := "disetujui"
			if status == "rejected" {
				statusText = "ditolak"
			}
			activities = append(activities, ActivityItem{
				Type:        "english",
				Description: fmt.Sprintf("Quest %s dari %s %s", title, name, statusText),
				Timestamp:   reviewedAt,
			})
		}
		rows.Close()

		// Get recent point claims (last 5)
		rows, _ = db.Query(`
			SELECT pc.status, pc.submitted_at, s.name, pr.name
			FROM point_claims pc
			JOIN students s ON pc.student_id = s.id
			JOIN point_rules pr ON pc.rule_id = pr.id
			WHERE pc.status IN ('approved', 'rejected')
			ORDER BY pc.reviewed_at DESC
			LIMIT 5
		`)
		for rows.Next() {
			var status, submittedAt, studentName, ruleName string
			rows.Scan(&status, &submittedAt, &studentName, &ruleName)
			statusText := "disetujui"
			if status == "rejected" {
				statusText = "ditolak"
			}
			activities = append(activities, ActivityItem{
				Type:        "claim",
				Description: fmt.Sprintf("Klaim poin '%s' dari %s %s", ruleName, studentName, statusText),
				Timestamp:   submittedAt,
			})
		}
		rows.Close()

		// Get recent announcements (last 3)
		rows, _ = db.Query(`
			SELECT title, created_at
			FROM announcements
			ORDER BY created_at DESC
			LIMIT 3
		`)
		for rows.Next() {
			var title, createdAt string
			rows.Scan(&title, &createdAt)
			activities = append(activities, ActivityItem{
				Type:        "announcement",
				Description: fmt.Sprintf("Pengumuman: %s", title),
				Timestamp:   createdAt,
			})
		}
		rows.Close()

		// Sort all activities by timestamp (most recent first)
		type actWithTime struct {
			activity ActivityItem
			parsed   time.Time
		}
		var sorted []actWithTime
		for _, act := range activities {
			if act.Timestamp == "" {
				continue
			}
			var t time.Time
			formats := []string{
				"2006-01-02 15:04:05",
				"2006-01-02T15:04:05Z",
				"2006-01-02T15:04:05",
				"2006-01-02",
			}
			for _, f := range formats {
				if parsed, err := time.Parse(f, act.Timestamp); err == nil {
					t = parsed
					break
				}
			}
			if t.IsZero() {
				continue
			}
			sorted = append(sorted, actWithTime{activity: act, parsed: t})
		}

		// Sort by time descending
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].parsed.After(sorted[i].parsed) {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		// Take top 20 and calculate time ago
		result := []ActivityItem{}
		for i := 0; i < len(sorted) && i < 20; i++ {
			act := sorted[i].activity
			act.TimeAgo = timeAgo(sorted[i].parsed)
			result = append(result, act)
		}

		return c.JSON(http.StatusOK, result)
	}
}

func timeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "baru saja"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		return fmt.Sprintf("%d menit lalu", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d jam lalu", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "kemarin"
		}
		return fmt.Sprintf("%d hari lalu", days)
	} else if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / 24 / 7)
		return fmt.Sprintf("%d minggu lalu", weeks)
	} else {
		return t.Format("02 Jan 2006")
	}
}
