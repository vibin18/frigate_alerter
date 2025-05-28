package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vibin/frigate_alerter/internal/adapters"
	"github.com/vibin/frigate_alerter/internal/application"
	"github.com/vibin/frigate_alerter/internal/config"
	"github.com/vibin/frigate_alerter/internal/domain"
	"github.com/vibin/frigate_alerter/internal/logger"
)

func main() {
	// Initialize logger with JSON formatting
	logger.Configure(logger.Config{
		Level: slog.LevelInfo,
		Output: os.Stdout,
		AddSource: true,
	})

	slog.Info("Starting Frigate Alerter service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	
	slog.Info("Configuration loaded successfully", "frigate_server", cfg.FrigateServer, "mqtt_server", cfg.MQTTServer, "time_zone", cfg.TimeZone)

	// Create SQLite repository
	repository, err := adapters.NewSQLiteAlertRepository("./alerts.db", cfg.Location)
	if err != nil {
		slog.Error("Failed to create SQLite repository", "error", err)
		os.Exit(1)
	}
	defer repository.Close()

	// Create Discord notifier
	notifier, err := adapters.NewDiscordNotifier(
		cfg.DiscordToken,
		cfg.DiscordChannelID,
		cfg.FrigateServer,
		cfg.FrigatePort,
	)
	if err != nil {
		slog.Error("Failed to create Discord notifier", "error", err)
		os.Exit(1)
	}
	defer notifier.Close()

	// Create alert service
	alertService := application.NewAlertService(repository, notifier, cfg)

	// Create MQTT subscriber
	subscriber, err := adapters.NewMQTTSubscriber(cfg.MQTTServer)
	if err != nil {
		slog.Error("Failed to create MQTT subscriber", "error", err)
		os.Exit(1)
	}
	defer subscriber.Close()

	// Subscribe to Frigate events
	err = subscriber.Subscribe(func(event *domain.FrigateEvent) {
		if err := alertService.ProcessEvent(event); err != nil {
			slog.Error("Error processing event", "error", err, "event_type", event.Type)
		}
	})
	if err != nil {
		slog.Error("Failed to subscribe to MQTT topic", "error", err)
		os.Exit(1)
	}

	slog.Info("Frigate Alerter service started successfully")
	slog.Info("Listening for events", "mqtt_server", cfg.MQTTServer)

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down Frigate Alerter service")
}
