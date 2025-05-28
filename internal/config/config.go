package config

import (
	"encoding/json"
	"os"
	"time"
)

// Config holds the application configuration
type Config struct {
	FrigateServer    string `json:"frigate_server"`
	FrigatePort      string `json:"frigate_port"`
	MQTTServer       string `json:"mqtt_server"`
	DiscordToken     string `json:"discord_token"`
	DiscordChannelID string `json:"discord_channel_id"`
	TimeZone         string `json:"time_zone"`
	ServerPort       string `json:"server_port"`
	Location         *time.Location
}

// LoadConfig loads configuration from environment variables and config.json file
func LoadConfig() (*Config, error) {
	config := &Config{
		FrigateServer:    getEnv("FRIGATE_SERVER", "localhost"),
		FrigatePort:      getEnv("FRIGATE_PORT", "5000"),
		MQTTServer:       getEnv("MQTT_SERVER", "tcp://localhost:1883"),
		DiscordToken:     getEnv("DISCORD_TOKEN", ""),
		DiscordChannelID: getEnv("DISCORD_CHANNEL_ID", ""),
		TimeZone:         getEnv("TIME_ZONE", "UTC"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
	}

	// Try to load from config.json if it exists
	if _, err := os.Stat("config.json"); err == nil {
		file, err := os.Open("config.json")
		if err != nil {
			return nil, err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			return nil, err
		}
	}

	// Set the time location
	location, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		return nil, err
	}
	config.Location = location

	return config, nil
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
