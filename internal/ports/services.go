package ports

import (
	"github.com/vibin/frigate_alerter/internal/domain"
)

// AlertNotifier defines the interface for sending notifications
type AlertNotifier interface {
	// SendAlert sends an alert notification
	SendAlert(alert *domain.Alert) error
}

// EventSubscriber defines the interface for subscribing to events
type EventSubscriber interface {
	// Subscribe starts listening for events
	Subscribe(handler func(event *domain.FrigateEvent)) error
	// Close stops listening for events
	Close() error
}

// AlertService defines the interface for alert business logic
type AlertService interface {
	// ProcessEvent processes a Frigate event and triggers alerts if needed
	ProcessEvent(event *domain.FrigateEvent) error
}
