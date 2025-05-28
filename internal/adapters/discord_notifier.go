package adapters

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/vibin/frigate_alerter/internal/domain"
)

// DiscordNotifier implements the AlertNotifier interface
type DiscordNotifier struct {
	session     *discordgo.Session
	channelID   string
	frigateURL  string
}

// NewDiscordNotifier creates a new Discord notifier
func NewDiscordNotifier(token string, channelID string, frigateServer string, frigatePort string) (*DiscordNotifier, error) {
	slog.Info("Initializing Discord notifier")
	
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		slog.Error("Failed to create Discord session", "error", err)
		return nil, err
	}

	// Open a websocket connection to Discord
	slog.Debug("Opening websocket connection to Discord")
	err = session.Open()
	if err != nil {
		slog.Error("Failed to open Discord connection", "error", err)
		return nil, err
	}

	frigateURL := fmt.Sprintf("http://%s:%s", frigateServer, frigatePort)
	slog.Info("Discord notifier initialized", "channel_id", channelID, "frigate_url", frigateURL)

	return &DiscordNotifier{
		session:     session,
		channelID:   channelID,
		frigateURL:  frigateURL,
	}, nil
}

// SendAlert sends an alert notification to Discord
func (d *DiscordNotifier) SendAlert(alert *domain.Alert) error {
	slog.Info("Sending alert to Discord", "camera", alert.CameraName, "alert_id", alert.ID)
	
	// Generate the image URL for this camera
	imageURL := fmt.Sprintf("%s/api/%s/latest.jpg?h=300&_t=%d", d.frigateURL, alert.CameraName, time.Now().Unix())
	slog.Info("Fetching image from Frigate", "url", imageURL)

	// Fetch the image from Frigate
	imageData, err := d.fetchImage(imageURL)
	if err != nil {
		slog.Error("Failed to fetch image from Frigate", "error", err, "url", imageURL)
		// Continue with the notification even if image fetch fails
	}

	// Create fields for additional info
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Camera",
			Value:  alert.CameraName,
			Inline: true,
		},
		{
			Name:   "Alert ID",
			Value:  alert.ID,
			Inline: true,
		},
		{
			Name:   "Time",
			Value:  alert.TriggeredAt.Format("2006-01-02 15:04:05"),
			Inline: true,
		},
	}

	// Create the message embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Alert from %s camera", alert.CameraName),
		Description: alert.AlertMessage,
		Color:       0xff0000, // Red color for alerts
		Fields:      fields,
		Timestamp:   alert.TriggeredAt.Format("2006-01-02T15:04:05-0700"),
	}

	// Prepare file name for the image
	imageFileName := fmt.Sprintf("%s_alert_%s.jpg", alert.CameraName, time.Now().Format("20060102_150405"))

	var sendErr error

	if imageData != nil {
		// Send the message with embed and file attachment
		file := &discordgo.File{
			Name:   imageFileName,
			Reader: bytes.NewReader(imageData),
		}

		messageData := &discordgo.MessageSend{
			Embed: embed,
			Files: []*discordgo.File{file},
		}

		_, sendErr = d.session.ChannelMessageSendComplex(d.channelID, messageData)
	} else {
		// If image fetch failed, just send the embed
		_, sendErr = d.session.ChannelMessageSendEmbed(d.channelID, embed)
	}

	if sendErr != nil {
		slog.Error("Failed to send Discord message", "error", sendErr, "channel_id", d.channelID)
		return sendErr
	}
	
	slog.Info("Successfully sent alert to Discord", "camera", alert.CameraName, "alert_id", alert.ID)
	return nil
}

// fetchImage downloads the image from the provided URL and returns it as a byte slice
func (d *DiscordNotifier) fetchImage(imageURL string) ([]byte, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Send GET request to fetch the image
	resp, err := client.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if response status is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	slog.Info("Successfully fetched image", "size_bytes", len(imageData))
	return imageData, nil
}

// Close closes the Discord session
func (d *DiscordNotifier) Close() error {
	slog.Info("Closing Discord session")
	return d.session.Close()
}
