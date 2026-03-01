package timer

import (
	"time"

	"jitrack/internal/db"
	"jitrack/internal/models"
)

// Timer manages the running timer state
type Timer struct {
	isRunning bool
	worklog   *models.Worklog
	ticker    *time.Ticker
	stopChan  chan bool
	db        *db.Database
	onTick    func(elapsedSeconds int)
	onStart   func(issueKey, summary string, startedAt time.Time)
	onStop    func(issueKey string, durationSeconds int)
}

// NewTimer creates a new Timer instance
func NewTimer(database *db.Database) *Timer {
	return &Timer{
		db:       database,
		stopChan: make(chan bool),
	}
}

// SetCallbacks sets the event callbacks
func (t *Timer) SetCallbacks(onTick func(int), onStart func(string, string, time.Time), onStop func(string, int)) {
	t.onTick = onTick
	t.onStart = onStart
	t.onStop = onStop
}

// StartTimer starts the timer for a task
func (t *Timer) StartTimer(issueKey string) (*models.Worklog, error) {
	// Stop any running timer first
	if t.isRunning {
		t.StopTimer()
	}
	
	// Start new timer in database
	worklog, err := t.db.StartTimer(issueKey)
	if err != nil {
		return nil, err
	}
	
	t.worklog = worklog
	t.isRunning = true
	t.ticker = time.NewTicker(time.Second)
	
	// Get task summary for the callback
	task, _ := t.db.GetTask(issueKey)
	summary := ""
	if task != nil {
		summary = task.Summary
	}
	
	// Emit start event
	if t.onStart != nil {
		t.onStart(issueKey, summary, worklog.StartedAt)
	}
	
	// Start ticker goroutine
	go func() {
		for {
			select {
			case <-t.ticker.C:
				if t.onTick != nil && t.worklog != nil {
					elapsed := t.worklog.ElapsedSeconds()
					t.onTick(elapsed)
				}
			case <-t.stopChan:
				return
			}
		}
	}()
	
	return worklog, nil
}

// StopTimer stops the timer
func (t *Timer) StopTimer() (*models.Worklog, error) {
	if !t.isRunning {
		return nil, nil
	}
	
	// Stop ticker
	if t.ticker != nil {
		t.ticker.Stop()
	}
	
	// Signal goroutine to stop
	select {
	case <-t.stopChan:
		// Channel already closed
	default:
		close(t.stopChan)
	}
	t.stopChan = make(chan bool)
	
	// Stop timer in database
	worklog, err := t.db.StopTimer()
	if err != nil {
		return nil, err
	}
	
	// Emit stop event
	if t.onStop != nil && worklog != nil && worklog.DurationSeconds != nil {
		t.onStop(worklog.IssueKey, *worklog.DurationSeconds)
	}
	
	t.worklog = nil
	t.isRunning = false
	
	return worklog, nil
}

// IsRunning returns whether the timer is currently running
func (t *Timer) IsRunning() bool {
	return t.isRunning
}

// GetRunningTimer returns the current running worklog
func (t *Timer) GetRunningTimer() *models.Worklog {
	// If we have a running timer in memory, return it
	if t.isRunning && t.worklog != nil {
		return t.worklog
	}
	
	// Otherwise check database for any running timer (e.g., from previous session)
	worklog, err := t.db.GetRunningTimer()
	if err != nil {
		return nil
	}
	
	if worklog != nil {
		t.worklog = worklog
		t.isRunning = true
		
		// Restart ticker
		t.ticker = time.NewTicker(time.Second)
		go func() {
			for {
				select {
				case <-t.ticker.C:
					if t.onTick != nil && t.worklog != nil {
						elapsed := t.worklog.ElapsedSeconds()
						t.onTick(elapsed)
					}
				case <-t.stopChan:
					return
				}
			}
		}()
	}
	
	return worklog
}

// GetTicker returns the ticker for external use
func (t *Timer) GetTicker() *time.Ticker {
	return t.ticker
}

// GetElapsedSeconds returns the elapsed time for the running timer
func (t *Timer) GetElapsedSeconds() int {
	if !t.isRunning || t.worklog == nil {
		return 0
	}
	return t.worklog.ElapsedSeconds()
}
