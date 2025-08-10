package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"pihole-analyzer/internal/logger"
)

// LogHandler handles log notifications
type LogHandler struct {
	logger *logger.Logger
}

// NewLogHandler creates a new log notification handler
func NewLogHandler(logger *logger.Logger) NotificationHandler {
	return &LogHandler{
		logger: logger.Component("log-notification"),
	}
}

// SendNotification logs the alert
func (h *LogHandler) SendNotification(ctx context.Context, alert *Alert, config interface{}) error {
	if alert == nil {
		h.logger.GetSlogger().Warn("Received nil alert in log notification")
		return nil // Handle gracefully
	}

	logLevel := slog.LevelInfo

	// Map alert severity to log level
	switch alert.Severity {
	case SeverityInfo:
		logLevel = slog.LevelInfo
	case SeverityWarning:
		logLevel = slog.LevelWarn
	case SeverityCritical, SeverityError:
		logLevel = slog.LevelError
	}

	// Create structured log entry
	logAttrs := []slog.Attr{
		slog.String("alert_id", alert.ID),
		slog.String("type", string(alert.Type)),
		slog.String("severity", string(alert.Severity)),
		slog.String("source", alert.Source),
		slog.Time("timestamp", alert.Timestamp),
	}

	if alert.ClientIP != "" {
		logAttrs = append(logAttrs, slog.String("client_ip", alert.ClientIP))
	}
	if alert.Domain != "" {
		logAttrs = append(logAttrs, slog.String("domain", alert.Domain))
	}
	if len(alert.Tags) > 0 {
		logAttrs = append(logAttrs, slog.String("tags", strings.Join(alert.Tags, ",")))
	}

	// Log with appropriate level
	h.logger.GetSlogger().LogAttrs(ctx, logLevel,
		fmt.Sprintf("🚨 ALERT: %s - %s", alert.Title, alert.Description),
		logAttrs...)

	return nil
}

// GetChannelType returns the channel type
func (h *LogHandler) GetChannelType() NotificationChannel {
	return ChannelLog
}

// TestNotification tests the log notification
func (h *LogHandler) TestNotification(ctx context.Context, config interface{}) error {
	h.logger.Info("Log notification test successful")
	return nil
}

// SlackHandler handles Slack notifications
type SlackHandler struct {
	config SlackConfig
	logger *logger.Logger
	client *http.Client
}

// NewSlackHandler creates a new Slack notification handler
func NewSlackHandler(config SlackConfig, logger *logger.Logger) NotificationHandler {
	return &SlackHandler{
		config: config,
		logger: logger.Component("slack-notification"),
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// SlackMessage represents a Slack webhook message
type SlackMessage struct {
	Channel     string            `json:"channel,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Text        string            `json:"text,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	Text      string       `json:"text,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

// SlackField represents a Slack attachment field
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// SendNotification sends alert to Slack
func (h *SlackHandler) SendNotification(ctx context.Context, alert *Alert, config interface{}) error {
	if h.config.WebhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}

	// Create Slack message
	message := h.createSlackMessage(alert)

	// Send to Slack
	return h.sendToSlack(ctx, message)
}

// createSlackMessage creates a formatted Slack message from an alert
func (h *SlackHandler) createSlackMessage(alert *Alert) SlackMessage {
	// Map severity to color
	color := "#36a64f" // green (info)
	switch alert.Severity {
	case SeverityWarning:
		color = "#ffaa00" // orange
	case SeverityError:
		color = "#ff0000" // red
	case SeverityCritical:
		color = "#800080" // purple
	}

	// Create fields
	fields := []SlackField{
		{Title: "Severity", Value: string(alert.Severity), Short: true},
		{Title: "Type", Value: string(alert.Type), Short: true},
		{Title: "Source", Value: alert.Source, Short: true},
		{Title: "Timestamp", Value: alert.Timestamp.Format("2006-01-02 15:04:05"), Short: true},
	}

	if alert.ClientIP != "" {
		fields = append(fields, SlackField{Title: "Client IP", Value: alert.ClientIP, Short: true})
	}
	if alert.Domain != "" {
		fields = append(fields, SlackField{Title: "Domain", Value: alert.Domain, Short: true})
	}
	if len(alert.Tags) > 0 {
		fields = append(fields, SlackField{Title: "Tags", Value: strings.Join(alert.Tags, ", "), Short: false})
	}

	attachment := SlackAttachment{
		Color:     color,
		Title:     alert.Title,
		Text:      alert.Description,
		Fields:    fields,
		Timestamp: alert.Timestamp.Unix(),
	}

	emoji := h.config.IconEmoji
	if emoji == "" {
		emoji = ":warning:"
	}

	return SlackMessage{
		Channel:     h.config.Channel,
		Username:    h.config.Username,
		IconEmoji:   emoji,
		Text:        fmt.Sprintf("🚨 *Alert:* %s", alert.Title),
		Attachments: []SlackAttachment{attachment},
	}
}

// sendToSlack sends the message to Slack webhook
func (h *SlackHandler) sendToSlack(ctx context.Context, message SlackMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API returned status %d", resp.StatusCode)
	}

	h.logger.GetSlogger().Debug("Slack notification sent successfully", slog.String("channel", h.config.Channel))
	return nil
}

// GetChannelType returns the channel type
func (h *SlackHandler) GetChannelType() NotificationChannel {
	return ChannelSlack
}

// TestNotification tests the Slack notification
func (h *SlackHandler) TestNotification(ctx context.Context, config interface{}) error {
	testAlert := &Alert{
		ID:          "test-alert",
		Type:        AlertTypeConfiguration,
		Severity:    SeverityInfo,
		Status:      AlertStatusFired,
		Title:       "Test Alert",
		Description: "This is a test alert to verify Slack integration",
		Timestamp:   time.Now(),
		Source:      "test",
		Tags:        []string{"test"},
	}

	return h.SendNotification(ctx, testAlert, config)
}

// EmailHandler handles email notifications
type EmailHandler struct {
	config EmailConfig
	logger *logger.Logger
}

// NewEmailHandler creates a new email notification handler
func NewEmailHandler(config EmailConfig, logger *logger.Logger) NotificationHandler {
	return &EmailHandler{
		config: config,
		logger: logger.Component("email-notification"),
	}
}

// SendNotification sends alert via email
func (h *EmailHandler) SendNotification(ctx context.Context, alert *Alert, config interface{}) error {
	if h.config.SMTPHost == "" || h.config.From == "" {
		return fmt.Errorf("email configuration incomplete")
	}

	if len(h.config.Recipients) == 0 {
		return fmt.Errorf("no email recipients configured")
	}

	// Create email content
	subject := fmt.Sprintf("[PiHole Alert] %s - %s", alert.Severity, alert.Title)
	body := h.createEmailBody(alert)

	// Send email
	return h.sendEmail(subject, body, h.config.Recipients)
}

// createEmailBody creates the email body for an alert
func (h *EmailHandler) createEmailBody(alert *Alert) string {
	var body strings.Builder

	body.WriteString(fmt.Sprintf("Alert Details:\n"))
	body.WriteString(fmt.Sprintf("==============\n\n"))
	body.WriteString(fmt.Sprintf("ID: %s\n", alert.ID))
	body.WriteString(fmt.Sprintf("Title: %s\n", alert.Title))
	body.WriteString(fmt.Sprintf("Description: %s\n", alert.Description))
	body.WriteString(fmt.Sprintf("Severity: %s\n", alert.Severity))
	body.WriteString(fmt.Sprintf("Type: %s\n", alert.Type))
	body.WriteString(fmt.Sprintf("Status: %s\n", alert.Status))
	body.WriteString(fmt.Sprintf("Source: %s\n", alert.Source))
	body.WriteString(fmt.Sprintf("Timestamp: %s\n", alert.Timestamp.Format("2006-01-02 15:04:05 MST")))

	if alert.ClientIP != "" {
		body.WriteString(fmt.Sprintf("Client IP: %s\n", alert.ClientIP))
	}
	if alert.Domain != "" {
		body.WriteString(fmt.Sprintf("Domain: %s\n", alert.Domain))
	}
	if len(alert.Tags) > 0 {
		body.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(alert.Tags, ", ")))
	}

	if alert.Anomaly != nil {
		body.WriteString(fmt.Sprintf("\nML Anomaly Details:\n"))
		body.WriteString(fmt.Sprintf("==================\n"))
		body.WriteString(fmt.Sprintf("Anomaly ID: %s\n", alert.Anomaly.ID))
		body.WriteString(fmt.Sprintf("Anomaly Type: %s\n", alert.Anomaly.Type))
		body.WriteString(fmt.Sprintf("ML Score: %.2f\n", alert.MLScore))
		body.WriteString(fmt.Sprintf("Confidence: %.2f\n", alert.Confidence))
	}

	body.WriteString(fmt.Sprintf("\n--\n"))
	body.WriteString(fmt.Sprintf("This alert was generated by PiHole Network Analyzer\n"))
	body.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05 MST")))

	return body.String()
}

// sendEmail sends an email using SMTP
func (h *EmailHandler) sendEmail(subject, body string, recipients []string) error {
	// Create message
	msg := h.createEmailMessage(subject, body, recipients)

	// Setup authentication
	auth := smtp.PlainAuth("", h.config.Username, h.config.Password, h.config.SMTPHost)

	// Setup server address
	addr := fmt.Sprintf("%s:%d", h.config.SMTPHost, h.config.SMTPPort)

	// Send email
	var err error
	if h.config.UseTLS {
		err = h.sendEmailTLS(addr, auth, msg, recipients)
	} else {
		err = smtp.SendMail(addr, auth, h.config.From, recipients, msg)
	}

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	h.logger.GetSlogger().Debug("Email notification sent successfully",
		slog.Int("recipients", len(recipients)),
		slog.String("smtp_host", h.config.SMTPHost))
	return nil
}

// sendEmailTLS sends email using TLS
func (h *EmailHandler) sendEmailTLS(addr string, auth smtp.Auth, msg []byte, recipients []string) error {
	// This is a simplified TLS implementation
	// In production, you might want to use a more robust SMTP library
	return smtp.SendMail(addr, auth, h.config.From, recipients, msg)
}

// createEmailMessage creates the email message
func (h *EmailHandler) createEmailMessage(subject, body string, recipients []string) []byte {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", h.config.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(recipients, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	return []byte(msg.String())
}

// GetChannelType returns the channel type
func (h *EmailHandler) GetChannelType() NotificationChannel {
	return ChannelEmail
}

// TestNotification tests the email notification
func (h *EmailHandler) TestNotification(ctx context.Context, config interface{}) error {
	testAlert := &Alert{
		ID:          "test-alert",
		Type:        AlertTypeConfiguration,
		Severity:    SeverityInfo,
		Status:      AlertStatusFired,
		Title:       "Test Alert",
		Description: "This is a test alert to verify email integration",
		Timestamp:   time.Now(),
		Source:      "test",
		Tags:        []string{"test"},
	}

	return h.SendNotification(ctx, testAlert, config)
}
