package domain

import (
	"time"
)

// Alert represents a detection alert from Frigate
type Alert struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	CameraName      string    `json:"camera_name"`
	TriggeredAt     time.Time `json:"triggered_at"`
	AlertMessage    string    `json:"alert_message"`
}

// FrigateEvent represents the event data received from MQTT
type FrigateEvent struct {
	Type   string      `json:"type"`
	Before FrigateBefore `json:"before"`
}

// FrigateBefore represents the "before" data in the Frigate event
type FrigateBefore struct {
	ID       string          `json:"id"`
	Camera   string          `json:"camera"`
	FrameTime float64         `json:"frame_time"`
	Snapshot  FrigateSnapshot `json:"snapshot"`
}

// FrigateSnapshot represents the snapshot data in the Frigate event
type FrigateSnapshot struct {
	FrameTime float64   `json:"frame_time"`
	Box       []int     `json:"box"`
	Area      int       `json:"area"`
	Region    []int     `json:"region"`
	Score     float64   `json:"score"`
}
