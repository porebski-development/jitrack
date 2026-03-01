package models

import (
	"fmt"
	"time"
)

// Task represents a JIRA task
type Task struct {
	IssueKey       string     `json:"issue_key"`
	Summary        string     `json:"summary"`
	Project        string     `json:"project"`
	URL            string     `json:"url"`
	EstimatedHours *float64   `json:"estimated_hours,omitempty"`
	IsCompleted    bool       `json:"is_completed"`
	ImportedAt     time.Time  `json:"imported_at"`
	LastUpdated    time.Time  `json:"last_updated"`
}

// Worklog represents a time tracking session
type Worklog struct {
	ID              *int       `json:"id,omitempty"`
	IssueKey        string     `json:"issue_key"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`
	IsRunning       bool       `json:"is_running"`
	Notes           *string    `json:"notes,omitempty"`
}

// ReportRow represents a daily report entry
type ReportRow struct {
	IssueKey     string  `json:"issue_key"`
	Summary      string  `json:"summary"`
	Sessions     int     `json:"sessions"`
	TotalSeconds int     `json:"total_seconds"`
}

// DurationFormatted returns formatted duration HH:MM:SS
func (w *Worklog) DurationFormatted() string {
	var duration int
	if w.DurationSeconds != nil {
		duration = *w.DurationSeconds
	} else if w.IsRunning {
		duration = int(time.Since(w.StartedAt).Seconds())
	} else {
		return "00:00:00"
	}

	hours := duration / 3600
	minutes := (duration % 3600) / 60
	seconds := duration % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// ElapsedSeconds returns current elapsed time for running timer
func (w *Worklog) ElapsedSeconds() int {
	if w.IsRunning {
		return int(time.Since(w.StartedAt).Seconds())
	}
	if w.DurationSeconds != nil {
		return *w.DurationSeconds
	}
	return 0
}

// FormatDuration formats seconds to HH:MM:SS
func FormatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}
