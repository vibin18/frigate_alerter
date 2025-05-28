package adapters

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/vibin/frigate_alerter/internal/config"
	"github.com/vibin/frigate_alerter/internal/domain"
	"github.com/vibin/frigate_alerter/internal/ports"
)

// HTTPServer provides a web UI for the Frigate alerter
type HTTPServer struct {
	repository      ports.AlertRepository
	notifier        ports.AlertNotifier
	config          *config.Config
	frigateService  *FrigateService
	templatesDir    string
	staticDir       string
	server          *http.Server
}

// CameraInfo represents information about a camera
type CameraInfo struct {
	Name  string `json:"name"`
	Count int    `json:"alert_count"`
}

// AlertResponse represents the result of an on-demand alert
type AlertResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	AlertID string `json:"alert_id,omitempty"`
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(
	repository ports.AlertRepository,
	notifier ports.AlertNotifier,
	frigateService *FrigateService,
	config *config.Config,
) *HTTPServer {
	return &HTTPServer{
		repository:     repository,
		notifier:       notifier,
		config:         config,
		frigateService: frigateService,
		templatesDir:   "web/templates",
		staticDir:      "web/static",
	}
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	router := http.NewServeMux()

	// Static files handler
	fs := http.FileServer(http.Dir(s.staticDir))
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	// Web UI routes
	router.HandleFunc("/", s.handleHome)
	router.HandleFunc("/cameras", s.handleCameras)
	router.HandleFunc("/alerts", s.handleAlerts)
	router.HandleFunc("/camera/", s.handleCameraDetails)

	// API routes
	router.HandleFunc("/api/cameras", s.handleAPIGetCameras)
	router.HandleFunc("/api/alerts", s.handleAPIGetAlerts)
	router.HandleFunc("/api/trigger", s.handleAPITriggerSnapshot)

	addr := fmt.Sprintf(":%s", s.config.ServerPort)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("Starting HTTP server", "address", addr)
	return s.server.ListenAndServe()
}

// Stop stops the HTTP server
func (s *HTTPServer) Stop() error {
	if s.server != nil {
		slog.Info("Stopping HTTP server")
		return s.server.Close()
	}
	return nil
}

// handleHome handles the home page request
func (s *HTTPServer) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles(
		filepath.Join(s.templatesDir, "layout.html"),
		filepath.Join(s.templatesDir, "home.html"),
	)
	if err != nil {
		slog.Error("Failed to parse template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
	}{
		Title: "Frigate Alerter - Home",
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		slog.Error("Failed to render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleCameras handles the cameras page request
func (s *HTTPServer) handleCameras(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		filepath.Join(s.templatesDir, "layout.html"),
		filepath.Join(s.templatesDir, "cameras.html"),
	)
	if err != nil {
		slog.Error("Failed to parse template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get camera info
	cameras, err := s.frigateService.GetCameras()
	if err != nil {
		slog.Error("Failed to get cameras", "error", err)
		http.Error(w, "Failed to get cameras", http.StatusInternalServerError)
		return
	}

	// Create camera data structures with config info included
	cameraData := make([]map[string]interface{}, 0, len(cameras))
	for _, camera := range cameras {
		cameraData = append(cameraData, map[string]interface{}{
			"Camera":        camera,
			"FrigateServer": s.config.FrigateServer,
			"FrigatePort":   s.config.FrigatePort,
		})
	}

	data := struct {
		Title   string
		Cameras []map[string]interface{}
		Config  *config.Config
	}{
		Title:   "Frigate Alerter - Cameras",
		Cameras: cameraData,
		Config:  s.config,
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		slog.Error("Failed to render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleAlerts handles the alerts page request
func (s *HTTPServer) handleAlerts(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		filepath.Join(s.templatesDir, "layout.html"),
		filepath.Join(s.templatesDir, "alerts.html"),
	)
	if err != nil {
		slog.Error("Failed to parse template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Parse query parameters
	limit := 100
	offset := 0
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get alerts
	alerts, err := s.repository.GetAlerts(limit, offset)
	if err != nil {
		slog.Error("Failed to get alerts", "error", err)
		http.Error(w, "Failed to get alerts", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title  string
		Alerts []*domain.Alert
		Config *config.Config
	}{
		Title:  "Frigate Alerter - Alerts",
		Alerts: alerts,
		Config: s.config,
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		slog.Error("Failed to render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleCameraDetails handles requests for a specific camera's details
func (s *HTTPServer) handleCameraDetails(w http.ResponseWriter, r *http.Request) {
	cameraName := r.URL.Path[len("/camera/"):]
	if cameraName == "" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles(
		filepath.Join(s.templatesDir, "layout.html"),
		filepath.Join(s.templatesDir, "camera_details.html"),
	)
	if err != nil {
		slog.Error("Failed to parse template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Parse query parameters
	limit := 50
	offset := 0
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get alerts for this camera
	alerts, err := s.repository.GetAlertsByCameraName(cameraName, limit, offset)
	if err != nil {
		slog.Error("Failed to get alerts for camera", "error", err, "camera", cameraName)
		http.Error(w, "Failed to get alerts", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title      string
		CameraName string
		Alerts     []*domain.Alert
		Config     *config.Config
	}{
		Title:      fmt.Sprintf("Frigate Alerter - %s Camera", cameraName),
		CameraName: cameraName,
		Alerts:     alerts,
		Config:     s.config,
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		slog.Error("Failed to render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// API Handlers

// handleAPIGetCameras returns a list of cameras as JSON
func (s *HTTPServer) handleAPIGetCameras(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cameras, err := s.frigateService.GetCameras()
	if err != nil {
		slog.Error("Failed to get cameras", "error", err)
		http.Error(w, `{"error":"Failed to get cameras"}`, http.StatusInternalServerError)
		return
	}

	// Get alert counts for each camera
	cameraInfos := make([]CameraInfo, 0, len(cameras))
	for _, camera := range cameras {
		alerts, err := s.repository.GetAlertsByCameraName(camera, 1, 0)
		if err != nil {
			slog.Error("Failed to get alert count for camera", "error", err, "camera", camera)
			continue
		}
		cameraInfos = append(cameraInfos, CameraInfo{
			Name:  camera,
			Count: len(alerts),
		})
	}

	if err := json.NewEncoder(w).Encode(cameraInfos); err != nil {
		slog.Error("Failed to encode cameras", "error", err)
		http.Error(w, `{"error":"Failed to encode cameras"}`, http.StatusInternalServerError)
	}
}

// handleAPIGetAlerts returns alerts as JSON
func (s *HTTPServer) handleAPIGetAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	limit := 100
	offset := 0
	camera := r.URL.Query().Get("camera")

	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	var alerts []*domain.Alert
	var err error

	if camera != "" {
		// Get alerts for a specific camera
		alerts, err = s.repository.GetAlertsByCameraName(camera, limit, offset)
	} else {
		// Get all alerts
		alerts, err = s.repository.GetAlerts(limit, offset)
	}

	if err != nil {
		slog.Error("Failed to get alerts", "error", err)
		http.Error(w, `{"error":"Failed to get alerts"}`, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		slog.Error("Failed to encode alerts", "error", err)
		http.Error(w, `{"error":"Failed to encode alerts"}`, http.StatusInternalServerError)
	}
}

// handleAPITriggerSnapshot handles requests to trigger a snapshot and send to Discord
func (s *HTTPServer) handleAPITriggerSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	// Parse JSON request
	var requestBody struct {
		Camera string `json:"camera"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, `{"success":false,"message":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	
	if requestBody.Camera == "" {
		http.Error(w, `{"success":false,"message":"Camera name is required"}`, http.StatusBadRequest)
		return
	}
	
	// Create manual alert
	currentTime := time.Now().In(s.config.Location)
	alertID := fmt.Sprintf("manual_%s_%d", requestBody.Camera, currentTime.UnixNano())
	
	alert := &domain.Alert{
		ID:           alertID,
		Type:         "manual",
		CameraName:   requestBody.Camera,
		TriggeredAt:  currentTime,
		AlertMessage: fmt.Sprintf("Manual snapshot from %s camera", requestBody.Camera),
	}
	
	// Save alert to database
	if err := s.repository.SaveAlert(alert); err != nil {
		slog.Error("Failed to save manual alert", "error", err, "camera", requestBody.Camera)
		http.Error(w, `{"success":false,"message":"Failed to save alert"}`, http.StatusInternalServerError)
		return
	}
	
	// Send alert to Discord
	if err := s.notifier.SendAlert(alert); err != nil {
		slog.Error("Failed to send manual alert to Discord", "error", err, "camera", requestBody.Camera)
		// Continue even if Discord notification fails
	}
	
	response := AlertResponse{
		Success: true,
		Message: fmt.Sprintf("Manual snapshot from %s camera sent to Discord", requestBody.Camera),
		AlertID: alertID,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, `{"success":false,"message":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}
