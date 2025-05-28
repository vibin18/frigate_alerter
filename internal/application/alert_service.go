package application

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/vibin/frigate_alerter/internal/config"
	"github.com/vibin/frigate_alerter/internal/domain"
	"github.com/vibin/frigate_alerter/internal/ports"
)

// AlertService implements the AlertService interface
type AlertService struct {
	repository ports.AlertRepository
	notifier   ports.AlertNotifier
	config     *config.Config
}

// NewAlertService creates a new alert service
func NewAlertService(
	repository ports.AlertRepository,
	notifier ports.AlertNotifier,
	config *config.Config,
) *AlertService {
	return &AlertService{
		repository: repository,
		notifier:   notifier,
		config:     config,
	}
}

// ProcessEvent processes a Frigate event and triggers alerts if needed
func (s *AlertService) ProcessEvent(event *domain.FrigateEvent) error {
	// Check if this is a "new" alert
	if event.Type != "new" {
		slog.Debug("Ignoring non-new event", "type", event.Type)
		return nil
	}

	// Create the alert message
	alertMessage := fmt.Sprintf("An object detected in the %s camera", event.Before.Camera)

	// Create a unique ID for this alert by combining the event ID with the camera name and current timestamp
	// This ensures we don't get primary key conflicts when duplicate MQTT messages are received
	currentTime := time.Now().In(s.config.Location)
	uniqueID := fmt.Sprintf("%s_%s_%d", event.Before.ID, event.Before.Camera, currentTime.UnixNano())
	
	// Create the alert object
	alert := &domain.Alert{
		ID:           uniqueID,
		Type:         event.Type,
		CameraName:   event.Before.Camera,
		TriggeredAt:  time.Now().In(s.config.Location),
		AlertMessage: alertMessage,
	}

	// Save alert to the database
	if err := s.repository.SaveAlert(alert); err != nil {
		slog.Error("Failed to save alert to database", "error", err, "camera", alert.CameraName, "alert_id", alert.ID)
		return err
	}

	// Send alert notification
	if err := s.notifier.SendAlert(alert); err != nil {
		slog.Error("Failed to send alert notification", "error", err, "camera", alert.CameraName, "alert_id", alert.ID)
		return err
	}

	slog.Info("Successfully processed alert", "camera", alert.CameraName, "alert_id", alert.ID, "time", alert.TriggeredAt)
	return nil
}
