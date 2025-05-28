package ports

import (
	"github.com/vibin/frigate_alerter/internal/domain"
)

// AlertRepository defines the interface for alert data storage operations
type AlertRepository interface {
	// SaveAlert saves an alert to the database
	SaveAlert(alert *domain.Alert) error
	
	// GetAlerts retrieves alerts based on optional filters
	GetAlerts(limit int, offset int) ([]*domain.Alert, error)
	
	// GetAlertsByCameraName retrieves alerts for a specific camera
	GetAlertsByCameraName(cameraName string, limit int, offset int) ([]*domain.Alert, error)
}
