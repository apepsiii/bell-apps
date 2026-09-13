package repository

import "database/sql"

// MigrationSemesterScores creates tables for the new composite scoring system
func MigrationSemesterScores(db *sql.DB) error {
	migrations := []string{
		// 1. Create semesters table
		`CREATE TABLE IF NOT EXISTS semesters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			academic_year TEXT NOT NULL,
			start_date DATE NOT NULL,
			end_date DATE NOT NULL,
			is_active BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 2. Create semester_scores table
		`CREATE TABLE IF NOT EXISTS semester_scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			semester_id INTEGER NOT NULL,
			attendance_index DECIMAL(5,2) DEFAULT 0,
			attendance_days INTEGER DEFAULT 0,
			late_days INTEGER DEFAULT 0,
			absent_days INTEGER DEFAULT 0,
			streak_days INTEGER DEFAULT 0,
			streak_bonus DECIMAL(5,2) DEFAULT 0,
			achievement_points INTEGER DEFAULT 0,
			achievement_normalized DECIMAL(5,2) DEFAULT 0,
			violation_points INTEGER DEFAULT 0,
			violation_burden DECIMAL(5,2) DEFAULT 0,
			redemption_points INTEGER DEFAULT 0,
			composite_score DECIMAL(5,2) DEFAULT 0,
			calculated_at DATETIME,
			FOREIGN KEY (student_id) REFERENCES students(id),
			FOREIGN KEY (semester_id) REFERENCES semesters(id),
			UNIQUE(student_id, semester_id)
		)`,

		// 3. Create attendance_streaks table for tracking consecutive attendance
		`CREATE TABLE IF NOT EXISTS attendance_streaks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			semester_id INTEGER NOT NULL,
			current_streak INTEGER DEFAULT 0,
			longest_streak INTEGER DEFAULT 0,
			last_attendance_date DATE,
			last_calculated_at DATETIME,
			FOREIGN KEY (student_id) REFERENCES students(id),
			FOREIGN KEY (semester_id) REFERENCES semesters(id),
			UNIQUE(student_id, semester_id)
		)`,

		// 4. Create violation_redemptions table
		`CREATE TABLE IF NOT EXISTS violation_redemptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			semester_id INTEGER NOT NULL,
			violation_id INTEGER NOT NULL,
			redemption_type TEXT NOT NULL,
			redemption_description TEXT,
			points_redeemed INTEGER NOT NULL,
			verified_by TEXT,
			verified_at DATETIME,
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (student_id) REFERENCES students(id),
			FOREIGN KEY (semester_id) REFERENCES semesters(id)
		)`,

		// 5. Create indexes
		`CREATE INDEX IF NOT EXISTS idx_semester_scores_student 
		 ON semester_scores(student_id)`,
		
		`CREATE INDEX IF NOT EXISTS idx_semester_scores_semester 
		 ON semester_scores(semester_id)`,
		
		`CREATE INDEX IF NOT EXISTS idx_semester_scores_composite 
		 ON semester_scores(composite_score DESC)`,

		`CREATE INDEX IF NOT EXISTS idx_attendance_streaks_student 
		 ON attendance_streaks(student_id, semester_id)`,

		`CREATE INDEX IF NOT EXISTS idx_violation_redemptions_student 
		 ON violation_redemptions(student_id, semester_id)`,

		// 6. Insert default semester if not exists
		`INSERT OR IGNORE INTO semesters (id, name, academic_year, start_date, end_date, is_active) VALUES
			(1, 'Semester Ganjil', '2026/2027', '2026-07-01', '2026-12-31', 1)`,

		// 7. Add config for composite scoring weights
		`INSERT OR IGNORE INTO point_config (config_key, config_value, description) VALUES
			('ATTENDANCE_WEIGHT', '30', 'Bobot Indeks Kehadiran (%)'),
			('ACHIEVEMENT_WEIGHT', '50', 'Bobot Poin Prestasi (%)'),
			('VIOLATION_WEIGHT', '20', 'Bobot Beban Pelanggaran (%)'),
			('STREAK_BONUS_DAYS', '10', 'Minimal hari berturut-turut untuk bonus'),
			('STREAK_BONUS_POINTS', '10', 'Bonus poin untuk streak'),
			('MAX_POINTS_PER_CATEGORY', '200', 'Max poin per kategori per semester')`,

		// 8. Add columns to pending_achievement_points
		`ALTER TABLE pending_achievement_points ADD COLUMN proof_url TEXT DEFAULT ''`,
		`ALTER TABLE pending_achievement_points ADD COLUMN semester_id INTEGER DEFAULT 1`,

		// 9. Create student_achievement_points table if not exists (for approved achievements)
		`CREATE TABLE IF NOT EXISTS student_achievement_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			rule_id INTEGER,
			pending_id INTEGER,
			points INTEGER NOT NULL,
			description TEXT,
			approved_by TEXT,
			approved_at DATETIME,
			semester_id INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			deleted_by TEXT,
			deletion_reason TEXT,
			FOREIGN KEY (student_id) REFERENCES students(id)
		)`,

		// 10. Create student_violation_points table if not exists (for Phase 3)
		`CREATE TABLE IF NOT EXISTS student_violation_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			rule_id INTEGER,
			points INTEGER NOT NULL,
			escalation_level INTEGER DEFAULT 1,
			description TEXT,
			recorded_by TEXT,
			recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			semester_id INTEGER DEFAULT 1,
			redemption_status TEXT DEFAULT 'none',
			redemption_points INTEGER DEFAULT 0,
			deleted_at DATETIME,
			deleted_by TEXT,
			deletion_reason TEXT,
			FOREIGN KEY (student_id) REFERENCES students(id)
		)`,

		// 11. Indexes for new tables
		`CREATE INDEX IF NOT EXISTS idx_student_achievement_points_student 
		 ON student_achievement_points(student_id, semester_id)`,
		
		`CREATE INDEX IF NOT EXISTS idx_student_violation_points_student 
		 ON student_violation_points(student_id, semester_id)`,
	}

	for _, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil {
			// Ignore duplicate/exists errors
			if err.Error() != "table semesters already exists" &&
				err.Error() != "table semester_scores already exists" &&
				err.Error() != "table attendance_streaks already exists" &&
				err.Error() != "table violation_redemptions already exists" {
				return err
			}
		}
	}

	return nil
}
