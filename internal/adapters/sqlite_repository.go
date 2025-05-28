package adapters

import (
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/vibin/frigate_alerter/internal/domain"
)

// SQLiteAlertRepository implements the AlertRepository interface using SQLite
type SQLiteAlertRepository struct {
	db       *sql.DB
	location *time.Location
}

// NewSQLiteAlertRepository creates a new SQLite repository
func NewSQLiteAlertRepository(dbPath string, location *time.Location) (*SQLiteAlertRepository, error) {
	slog.Info("Initializing SQLite repository", "path", dbPath)
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		slog.Error("Failed to open SQLite database", "path", dbPath, "error", err)
		return nil, err
	}

	repo := &SQLiteAlertRepository{
		db:       db,
		location: location,
	}

	if err := repo.initDB(); err != nil {
		slog.Error("Failed to initialize database schema", "error", err)
		db.Close()
		return nil, err
	}

	slog.Info("SQLite repository initialized successfully", "path", dbPath)
	return repo, nil
}

// initDB initializes the database schema
func (r *SQLiteAlertRepository) initDB() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			camera_name TEXT NOT NULL,
			triggered_at TIMESTAMP NOT NULL,
			alert_message TEXT NOT NULL
		)
	`)
	return err
}

// SaveAlert saves an alert to the database
func (r *SQLiteAlertRepository) SaveAlert(alert *domain.Alert) error {
	slog.Debug("Saving alert to database", "alert_id", alert.ID, "camera", alert.CameraName)
	
	_, err := r.db.Exec(
		`INSERT INTO alerts (id, type, camera_name, triggered_at, alert_message) 
		 VALUES (?, ?, ?, ?, ?)`,
		alert.ID,
		alert.Type,
		alert.CameraName,
		alert.TriggeredAt.In(r.location),
		alert.AlertMessage,
	)
	
	if err != nil {
		slog.Error("Failed to save alert to database", "alert_id", alert.ID, "error", err)
		return err
	}
	
	slog.Debug("Alert saved successfully", "alert_id", alert.ID)
	return nil
}

// GetAlerts retrieves alerts based on optional filters
func (r *SQLiteAlertRepository) GetAlerts(limit int, offset int) ([]*domain.Alert, error) {
	rows, err := r.db.Query(
		`SELECT id, type, camera_name, triggered_at, alert_message 
		 FROM alerts 
		 ORDER BY triggered_at DESC 
		 LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAlerts(rows)
}

// GetAlertsByCameraName retrieves alerts for a specific camera
func (r *SQLiteAlertRepository) GetAlertsByCameraName(cameraName string, limit int, offset int) ([]*domain.Alert, error) {
	rows, err := r.db.Query(
		`SELECT id, type, camera_name, triggered_at, alert_message 
		 FROM alerts 
		 WHERE camera_name = ? 
		 ORDER BY triggered_at DESC 
		 LIMIT ? OFFSET ?`,
		cameraName, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAlerts(rows)
}

// scanAlerts scans rows into alert objects
func (r *SQLiteAlertRepository) scanAlerts(rows *sql.Rows) ([]*domain.Alert, error) {
	var alerts []*domain.Alert
	for rows.Next() {
		var alert domain.Alert
		var triggeredAt string
		
		err := rows.Scan(
			&alert.ID,
			&alert.Type,
			&alert.CameraName,
			&triggeredAt,
			&alert.AlertMessage,
		)
		if err != nil {
			return nil, err
		}
		
		// Parse the timestamp - try multiple formats to handle different database outputs
		t, err := parseTime(triggeredAt)
		if err != nil {
			slog.Error("Failed to parse timestamp", "timestamp", triggeredAt, "error", err)
			return nil, err
		}
		alert.TriggeredAt = t
		
		alerts = append(alerts, &alert)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return alerts, nil
}

// parseTime attempts to parse a timestamp string in multiple formats
func parseTime(timeStr string) (time.Time, error) {
	// Try different time formats
	formats := []string{
		"2006-01-02 15:04:05.999999999-07:00", // SQLite default format
		"2006-01-02T15:04:05.999999999-07:00",  // ISO 8601 format
		"2006-01-02 15:04:05Z07:00",            // Simplified format
		"2006-01-02T15:04:05Z07:00",            // Another common format
		"2006-01-02 15:04:05",                  // Simple format without timezone
		"2006-01-02T15:04:05",                  // Simple ISO format
	}

	var firstErr error
	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	return time.Time{}, firstErr
}

// Close closes the database connection
func (r *SQLiteAlertRepository) Close() error {
	slog.Info("Closing SQLite database connection")
	return r.db.Close()
}
