package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"jitrack/internal/models"
	_ "modernc.org/sqlite"
)

// Database represents the SQLite database
type Database struct {
	conn *sql.DB
}

// New creates a new Database instance (alias for NewDatabase compatibility)
func New() (*Database, error) {
	return NewDatabase()
}

// NewDatabase creates a new Database instance
func NewDatabase() (*Database, error) {
	dbPath, err := GetDatabasePath()
	if err != nil {
		return nil, err
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &Database{conn: conn}
	if err := db.createTables(); err != nil {
		return nil, err
	}

	return db, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.conn.Close()
}

// GetDatabasePath returns the platform-specific database path
func GetDatabasePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// macOS
	return filepath.Join(homeDir, "Library", "Application Support", "JITRACK", "jitrack.db"), nil
}

// createTables creates the database schema and handles migrations
func (db *Database) createTables() error {
	// Create tasks table
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			issue_key TEXT PRIMARY KEY,
			summary TEXT NOT NULL,
			project TEXT NOT NULL,
			url TEXT,
			estimated_hours REAL,
			is_completed BOOLEAN DEFAULT 0,
			imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	// Create worklogs table
	_, err = db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS worklogs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			issue_key TEXT NOT NULL REFERENCES tasks(issue_key),
			started_at TIMESTAMP NOT NULL,
			ended_at TIMESTAMP,
			duration_seconds INTEGER,
			is_running BOOLEAN DEFAULT 0,
			notes TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create worklogs table: %w", err)
	}

	// Create indexes
	_, err = db.conn.Exec(`
		CREATE INDEX IF NOT EXISTS idx_worklogs_issue_key ON worklogs(issue_key);
		CREATE INDEX IF NOT EXISTS idx_worklogs_is_running ON worklogs(is_running)
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Migration: Add estimated_hours column if it doesn't exist (for old DBs)
	_, _ = db.conn.Exec(`ALTER TABLE tasks ADD COLUMN estimated_hours REAL`)
	
	// Migration: Add is_completed column if it doesn't exist (for old DBs)
	_, _ = db.conn.Exec(`ALTER TABLE tasks ADD COLUMN is_completed BOOLEAN DEFAULT 0`)

	return nil
}

// SaveTask saves or updates a task
func (db *Database) SaveTask(task models.Task) error {
	now := time.Now()
	
	// Check if task exists
	var exists bool
	err := db.conn.QueryRow("SELECT 1 FROM tasks WHERE issue_key = ?", task.IssueKey).Scan(&exists)
	
	if err == sql.ErrNoRows || !exists {
		// Insert new task
		_, err = db.conn.Exec(`
			INSERT INTO tasks (issue_key, summary, project, url, estimated_hours, is_completed, imported_at, last_updated)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, task.IssueKey, task.Summary, task.Project, task.URL, task.EstimatedHours, task.IsCompleted, now, now)
	} else {
		// Update existing task
		_, err = db.conn.Exec(`
			UPDATE tasks 
			SET summary = ?, project = ?, url = ?, estimated_hours = ?, is_completed = ?, last_updated = ?
			WHERE issue_key = ?
		`, task.Summary, task.Project, task.URL, task.EstimatedHours, task.IsCompleted, now, task.IssueKey)
	}
	
	return err
}

// GetTask retrieves a task by issue key
func (db *Database) GetTask(issueKey string) (*models.Task, error) {
	var task models.Task
	var estimatedHours sql.NullFloat64
	
	err := db.conn.QueryRow(`
		SELECT issue_key, summary, project, url, estimated_hours, is_completed, imported_at, last_updated
		FROM tasks WHERE issue_key = ?
	`, issueKey).Scan(
		&task.IssueKey, &task.Summary, &task.Project, &task.URL,
		&estimatedHours, &task.IsCompleted, &task.ImportedAt, &task.LastUpdated,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	if estimatedHours.Valid {
		task.EstimatedHours = &estimatedHours.Float64
	}
	
	return &task, nil
}

// GetAllTasks retrieves all tasks
func (db *Database) GetAllTasks() ([]models.Task, error) {
	rows, err := db.conn.Query(`
		SELECT issue_key, summary, project, url, estimated_hours, is_completed, imported_at, last_updated
		FROM tasks ORDER BY last_updated DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return scanTasks(rows)
}

// GetIncompleteTasks retrieves only incomplete tasks
func (db *Database) GetIncompleteTasks() ([]models.Task, error) {
	rows, err := db.conn.Query(`
		SELECT issue_key, summary, project, url, estimated_hours, is_completed, imported_at, last_updated
		FROM tasks WHERE is_completed = 0 ORDER BY last_updated DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return scanTasks(rows)
}

// scanTasks scans task rows into slice
func scanTasks(rows *sql.Rows) ([]models.Task, error) {
	var tasks []models.Task
	
	for rows.Next() {
		var task models.Task
		var estimatedHours sql.NullFloat64
		
		err := rows.Scan(
			&task.IssueKey, &task.Summary, &task.Project, &task.URL,
			&estimatedHours, &task.IsCompleted, &task.ImportedAt, &task.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		
		if estimatedHours.Valid {
			task.EstimatedHours = &estimatedHours.Float64
		}
		
		tasks = append(tasks, task)
	}
	
	return tasks, rows.Err()
}

// CompleteTask marks a task as completed
func (db *Database) CompleteTask(issueKey string) error {
	_, err := db.conn.Exec(
		"UPDATE tasks SET is_completed = 1, last_updated = ? WHERE issue_key = ?",
		time.Now(), issueKey,
	)
	return err
}

// UncompleteTask marks a task as not completed
func (db *Database) UncompleteTask(issueKey string) error {
	_, err := db.conn.Exec(
		"UPDATE tasks SET is_completed = 0, last_updated = ? WHERE issue_key = ?",
		time.Now(), issueKey,
	)
	return err
}

// DeleteTask deletes a task
func (db *Database) DeleteTask(issueKey string) error {
	// First delete associated worklogs
	_, err := db.conn.Exec("DELETE FROM worklogs WHERE issue_key = ?", issueKey)
	if err != nil {
		return err
	}
	
	// Then delete the task
	_, err = db.conn.Exec("DELETE FROM tasks WHERE issue_key = ?", issueKey)
	return err
}

// StartTimer starts a new timer for a task (stops any running timer first)
func (db *Database) StartTimer(issueKey string) (*models.Worklog, error) {
	// First stop any running timer
	if _, err := db.StopTimer(); err != nil {
		return nil, err
	}
	
	now := time.Now()
	
	result, err := db.conn.Exec(`
		INSERT INTO worklogs (issue_key, started_at, is_running)
		VALUES (?, ?, 1)
	`, issueKey, now)
	if err != nil {
		return nil, err
	}
	
	id, _ := result.LastInsertId()
	idInt := int(id)
	
	return &models.Worklog{
		ID:        &idInt,
		IssueKey:  issueKey,
		StartedAt: now,
		IsRunning: true,
	}, nil
}

// StopTimer stops the running timer
func (db *Database) StopTimer() (*models.Worklog, error) {
	// Get running timer
	var worklog models.Worklog
	var id int
	
	err := db.conn.QueryRow(`
		SELECT id, issue_key, started_at FROM worklogs 
		WHERE is_running = 1 LIMIT 1
	`).Scan(&id, &worklog.IssueKey, &worklog.StartedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	worklog.ID = &id
	
	// Stop the timer
	now := time.Now()
	duration := int(now.Sub(worklog.StartedAt).Seconds())
	
	_, err = db.conn.Exec(`
		UPDATE worklogs 
		SET ended_at = ?, duration_seconds = ?, is_running = 0
		WHERE id = ?
	`, now, duration, id)
	if err != nil {
		return nil, err
	}
	
	worklog.EndedAt = &now
	worklog.DurationSeconds = &duration
	worklog.IsRunning = false
	
	return &worklog, nil
}

// GetRunningTimer returns the currently running timer
func (db *Database) GetRunningTimer() (*models.Worklog, error) {
	var worklog models.Worklog
	var id int
	
	err := db.conn.QueryRow(`
		SELECT id, issue_key, started_at FROM worklogs 
		WHERE is_running = 1 LIMIT 1
	`).Scan(&id, &worklog.IssueKey, &worklog.StartedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	worklog.ID = &id
	worklog.IsRunning = true
	return &worklog, nil
}

// GetWorklogsForTask returns all worklogs for a task
func (db *Database) GetWorklogsForTask(issueKey string) ([]models.Worklog, error) {
	rows, err := db.conn.Query(`
		SELECT id, issue_key, started_at, ended_at, duration_seconds, is_running, notes
		FROM worklogs 
		WHERE issue_key = ?
		ORDER BY started_at DESC
	`, issueKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var worklogs []models.Worklog
	
	for rows.Next() {
		var w models.Worklog
		var id int
		var endedAt sql.NullTime
		var duration sql.NullInt64
		var notes sql.NullString
		
		err := rows.Scan(&id, &w.IssueKey, &w.StartedAt, &endedAt, &duration, &w.IsRunning, &notes)
		if err != nil {
			return nil, err
		}
		
		w.ID = &id
		if endedAt.Valid {
			w.EndedAt = &endedAt.Time
		}
		if duration.Valid {
			dur := int(duration.Int64)
			w.DurationSeconds = &dur
		}
		if notes.Valid {
			w.Notes = &notes.String
		}
		
		worklogs = append(worklogs, w)
	}
	
	return worklogs, rows.Err()
}

// GetDailyReport returns work summary for a specific date (YYYY-MM-DD format)
func (db *Database) GetDailyReport(date string) ([]models.ReportRow, error) {
	// Parse date to get start and end of day
	startOfDay, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	rows, err := db.conn.Query(`
		SELECT 
			w.issue_key,
			t.summary,
			COUNT(*) as sessions,
			COALESCE(SUM(w.duration_seconds), 0) as total_seconds
		FROM worklogs w
		JOIN tasks t ON w.issue_key = t.issue_key
		WHERE w.started_at >= ? AND w.started_at < ? AND w.duration_seconds IS NOT NULL
		GROUP BY w.issue_key, t.summary
		ORDER BY total_seconds DESC
	`, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var report []models.ReportRow
	
	for rows.Next() {
		var row models.ReportRow
		err := rows.Scan(&row.IssueKey, &row.Summary, &row.Sessions, &row.TotalSeconds)
		if err != nil {
			return nil, err
		}
		report = append(report, row)
	}
	
	return report, rows.Err()
}
