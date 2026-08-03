package repository

import "database/sql"

// MigrationAuditTrail adds audit trail tables and soft delete columns
func MigrationAuditTrail(db *sql.DB) error {
	migrations := []string{
		// 1. Create point_audit_log table
		`CREATE TABLE IF NOT EXISTS point_audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			action TEXT NOT NULL,
			table_name TEXT NOT NULL,
			record_id INTEGER NOT NULL,
			old_data TEXT,
			new_data TEXT,
			performed_by TEXT NOT NULL,
			performed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			ip_address TEXT,
			user_agent TEXT,
			reason TEXT
		)`,

		// 2. Create indexes for audit log
		`CREATE INDEX IF NOT EXISTS idx_audit_log_table_record 
		 ON point_audit_log(table_name, record_id)`,
		
		`CREATE INDEX IF NOT EXISTS idx_audit_log_performed_by 
		 ON point_audit_log(performed_by)`,
		
		`CREATE INDEX IF NOT EXISTS idx_audit_log_performed_at 
		 ON point_audit_log(performed_at)`,

		// 3. Add soft delete columns to student_achievement_points
		`ALTER TABLE student_achievement_points ADD COLUMN deleted_at DATETIME`,
		`ALTER TABLE student_achievement_points ADD COLUMN deleted_by TEXT`,
		`ALTER TABLE student_achievement_points ADD COLUMN deletion_reason TEXT`,

		// 4. Add soft delete columns to student_violation_points
		`ALTER TABLE student_violation_points ADD COLUMN deleted_at DATETIME`,
		`ALTER TABLE student_violation_points ADD COLUMN deleted_by TEXT`,
		`ALTER TABLE student_violation_points ADD COLUMN deletion_reason TEXT`,

		// 5. Create pending_achievement_points for approval workflow
		`CREATE TABLE IF NOT EXISTS pending_achievement_points (
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
			status TEXT DEFAULT 'pending',
			rejection_reason TEXT,
			FOREIGN KEY (student_id) REFERENCES students(id)
		)`,

		// 6. Create point_config table
		`CREATE TABLE IF NOT EXISTS point_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			config_key TEXT UNIQUE NOT NULL,
			config_value TEXT NOT NULL,
			description TEXT,
			updated_by TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 7. Insert default configs
		`INSERT OR IGNORE INTO point_config (config_key, config_value, description) VALUES
			('MAX_POINTS_PER_TRANSACTION', '100', 'Maksimal poin per transaksi'),
			('RATE_LIMIT_PER_DAY', '50', 'Maksimal input per hari per admin'),
			('APPROVAL_THRESHOLD', '50', 'Poin lebih dari ini perlu approval'),
			('SP1_THRESHOLD', '25', 'Threshold untuk SP1'),
			('SP2_THRESHOLD', '51', 'Threshold untuk SP2'),
			('SP3_THRESHOLD', '76', 'Threshold untuk SP3'),
			('CERT_THRESHOLD', '100', 'Threshold untuk sertifikat'),
			('REWARD_THRESHOLD', '126', 'Threshold untuk hadiah'),
			('WALUYA_THRESHOLD', '151', 'Threshold untuk Waluya Utama')`,
	}

	for _, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil {
			// Check if error is "duplicate column" which is okay
			if err.Error() != "duplicate column name: deleted_at" &&
				err.Error() != "duplicate column name: deleted_by" &&
				err.Error() != "duplicate column name: deletion_reason" {
				return err
			}
		}
	}

	return nil
}
