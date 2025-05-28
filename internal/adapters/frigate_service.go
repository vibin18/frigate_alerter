package adapters

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/vibin/frigate_alerter/internal/config"
)

// FrigateService provides methods for interacting with the Frigate API
type FrigateService struct {
	config *config.Config
	client *http.Client
}

// NewFrigateService creates a new Frigate service
func NewFrigateService(config *config.Config) *FrigateService {
	return &FrigateService{
		config: config,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// FrigateConfig represents the configuration from Frigate API
type FrigateConfig struct {
	Cameras map[string]interface{} `json:"cameras"`
}

// GetCameras returns a list of cameras from the Frigate API
func (s *FrigateService) GetCameras() ([]string, error) {
	url := fmt.Sprintf("%s/api/config", s.getBaseURL())
	
	slog.Debug("Fetching Frigate configuration", "url", url)
	
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to get Frigate config: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %w", err)
	}
	
	var config FrigateConfig
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal config: %w", err)
	}
	
	cameras := make([]string, 0, len(config.Cameras))
	for camera := range config.Cameras {
		cameras = append(cameras, camera)
	}
	
	slog.Debug("Found cameras", "count", len(cameras), "cameras", cameras)
	
	return cameras, nil
}

// GetSnapshot returns the latest snapshot for a camera
func (s *FrigateService) GetSnapshot(camera string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/%s/latest.jpg?h=300", s.getBaseURL(), camera)
	
	slog.Debug("Fetching snapshot", "camera", camera, "url", url)
	
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to get snapshot: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
	
	return io.ReadAll(resp.Body)
}

// getBaseURL returns the base URL for the Frigate API
func (s *FrigateService) getBaseURL() string {
	return fmt.Sprintf("http://%s:%s", s.config.FrigateServer, s.config.FrigatePort)
}
