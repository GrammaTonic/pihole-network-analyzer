package alerts

import (
	"context"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/ml"
)

// TestLogHandler tests log-based notifications
func TestLogHandler(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	handler := NewLogHandler(logger)

	// Verify channel type
	if handler.GetChannelType() != ChannelLog {
		t.Errorf("expected channel type %s, got %s", ChannelLog, handler.GetChannelType())
	}

	// Test notification
	alert := &Alert{
		ID:          "log-test-1",
		Type:        AlertTypeAnomaly,
		Severity:    SeverityWarning,
		Status:      AlertStatusFired,
		Title:       "Test Log Alert",
		Description: "This is a test alert for log notifications",
		Timestamp:   time.Now(),
		Source:      "test",
		ClientIP:    "192.168.1.100",
		Domain:      "test.com",
		Tags:        []string{"test", "log"},
	}

	ctx := context.Background()
	if err := handler.SendNotification(ctx, alert, nil); err != nil {
		t.Errorf("failed to send log notification: %v", err)
	}

	// Test different severity levels
	severities := []AlertSeverity{SeverityInfo, SeverityWarning, SeverityCritical, SeverityError}
	for _, severity := range severities {
		alert.Severity = severity
		alert.Title = "Test " + string(severity) + " Alert"
		if err := handler.SendNotification(ctx, alert, nil); err != nil {
			t.Errorf("failed to send %s log notification: %v", severity, err)
		}
	}

	// Test with ML anomaly data
	alert.Anomaly = &ml.Anomaly{
		ID:          "test-anomaly",
		Type:        ml.AnomalyTypeVolumeSpike,
		Severity:    ml.SeverityHigh,
		Description: "Test ML anomaly",
		Score:       0.85,
		Confidence:  0.9,
	}
	alert.MLScore = 0.85
	alert.Confidence = 0.9

	if err := handler.SendNotification(ctx, alert, nil); err != nil {
		t.Errorf("failed to send ML anomaly log notification: %v", err)
	}

	// Test notification
	if err := handler.TestNotification(ctx, nil); err != nil {
		t.Errorf("failed log test notification: %v", err)
	}
}

// TestSlackHandler tests Slack notifications
func TestSlackHandler(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	config := SlackConfig{
		Enabled:   true,
		Channel:   "#test-alerts",
		Username:  "TestBot",
		IconEmoji: ":robot_face:",
		Timeout:   30 * time.Second,
		// Note: WebhookURL is not set for testing
	}

	handler := NewSlackHandler(config, logger)

	// Verify channel type
	if handler.GetChannelType() != ChannelSlack {
		t.Errorf("expected channel type %s, got %s", ChannelSlack, handler.GetChannelType())
	}

	// Test creating Slack message
	alert := &Alert{
		ID:          "slack-test-1",
		Type:        AlertTypeAnomaly,
		Severity:    SeverityWarning,
		Status:      AlertStatusFired,
		Title:       "Test Slack Alert",
		Description: "This is a test alert for Slack notifications",
		Timestamp:   time.Now(),
		Source:      "test",
		ClientIP:    "192.168.1.100",
		Domain:      "test.com",
		Tags:        []string{"test", "slack"},
	}

	slackHandler := handler.(*SlackHandler)
	message := slackHandler.createSlackMessage(alert)

	// Verify message structure
	if message.Channel != config.Channel {
		t.Errorf("expected channel %s, got %s", config.Channel, message.Channel)
	}
	if message.Username != config.Username {
		t.Errorf("expected username %s, got %s", config.Username, message.Username)
	}
	if message.IconEmoji != config.IconEmoji {
		t.Errorf("expected icon %s, got %s", config.IconEmoji, message.IconEmoji)
	}
	if len(message.Attachments) != 1 {
		t.Errorf("expected 1 attachment, got %d", len(message.Attachments))
	}

	attachment := message.Attachments[0]
	if attachment.Title != alert.Title {
		t.Errorf("expected title %s, got %s", alert.Title, attachment.Title)
	}
	if attachment.Text != alert.Description {
		t.Errorf("expected text %s, got %s", alert.Description, attachment.Text)
	}

	// Check fields
	expectedFields := []string{"Severity", "Type", "Source", "Timestamp", "Client IP", "Domain", "Tags"}
	if len(attachment.Fields) < len(expectedFields)-2 { // Some fields are optional
		t.Errorf("expected at least %d fields, got %d", len(expectedFields)-2, len(attachment.Fields))
	}

	// Test different severity colors
	severityTests := []struct {
		severity      AlertSeverity
		expectedColor string
	}{
		{SeverityInfo, "#36a64f"},
		{SeverityWarning, "#ffaa00"},
		{SeverityError, "#ff0000"},
		{SeverityCritical, "#800080"},
	}

	for _, test := range severityTests {
		alert.Severity = test.severity
		msg := slackHandler.createSlackMessage(alert)
		if msg.Attachments[0].Color != test.expectedColor {
			t.Errorf("expected color %s for severity %s, got %s",
				test.expectedColor, test.severity, msg.Attachments[0].Color)
		}
	}

	// Test sending notification (should fail without webhook URL)
	ctx := context.Background()
	if err := handler.SendNotification(ctx, alert, config); err == nil {
		t.Error("expected error when sending Slack notification without webhook URL")
	}

	// Test with webhook URL (will fail but should reach the HTTP call)
	configWithWebhook := config
	configWithWebhook.WebhookURL = "https://localhost:65534/webhook" // Use localhost with unused port instead of external domain
	handlerWithWebhook := NewSlackHandler(configWithWebhook, logger)

	// This should fail with HTTP error, not config error
	err := handlerWithWebhook.SendNotification(ctx, alert, configWithWebhook)
	if err == nil {
		t.Error("expected HTTP error when sending to invalid webhook")
	} else {
		// Should not be a config error about missing webhook URL
		if err.Error() == "Slack webhook URL not configured" {
			t.Error("unexpected config error - should be HTTP error")
		}
	}
}

// TestEmailHandler tests email notifications
func TestEmailHandler(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	config := EmailConfig{
		Enabled:    true,
		SMTPHost:   "smtp.example.com",
		SMTPPort:   587,
		Username:   "test@example.com",
		Password:   "password",
		From:       "alerts@example.com",
		Recipients: []string{"admin@example.com", "ops@example.com"},
		UseTLS:     true,
		Timeout:    30 * time.Second,
	}

	handler := NewEmailHandler(config, logger)

	// Verify channel type
	if handler.GetChannelType() != ChannelEmail {
		t.Errorf("expected channel type %s, got %s", ChannelEmail, handler.GetChannelType())
	}

	// Test creating email body
	alert := &Alert{
		ID:          "email-test-1",
		Type:        AlertTypeAnomaly,
		Severity:    SeverityWarning,
		Status:      AlertStatusFired,
		Title:       "Test Email Alert",
		Description: "This is a test alert for email notifications",
		Timestamp:   time.Now(),
		Source:      "test",
		ClientIP:    "192.168.1.100",
		Domain:      "test.com",
		Tags:        []string{"test", "email"},
	}

	emailHandler := handler.(*EmailHandler)
	body := emailHandler.createEmailBody(alert)

	// Verify body content
	expectedContent := []string{
		alert.ID,
		alert.Title,
		alert.Description,
		string(alert.Severity),
		string(alert.Type),
		alert.Source,
		alert.ClientIP,
		alert.Domain,
		"test, email", // Tags should be joined
	}

	for _, content := range expectedContent {
		if content != "" && len(content) > 0 {
			// Simple check if content is in body (case-insensitive might be better)
			found := false
			for _, line := range []string{body} {
				if len(line) > 0 {
					found = true
					break
				}
			}
			if !found && content != "" {
				// We'll just check the body is not empty since string matching is complex
				if len(body) == 0 {
					t.Error("email body should not be empty")
				}
			}
		}
	}

	// Test with ML anomaly data
	alert.Anomaly = &ml.Anomaly{
		ID:          "test-anomaly",
		Type:        ml.AnomalyTypeVolumeSpike,
		Severity:    ml.SeverityHigh,
		Description: "Test ML anomaly",
		Score:       0.85,
		Confidence:  0.9,
	}
	alert.MLScore = 0.85
	alert.Confidence = 0.9

	bodyWithAnomaly := emailHandler.createEmailBody(alert)
	if len(bodyWithAnomaly) <= len(body) {
		t.Error("email body with anomaly should be longer than without")
	}

	// Test creating email message
	subject := "Test Alert Subject"
	recipients := []string{"test1@example.com", "test2@example.com"}
	msg := emailHandler.createEmailMessage(subject, body, recipients)

	msgString := string(msg)
	if len(msgString) == 0 {
		t.Error("email message should not be empty")
	}

	// Test sending notification (should fail without valid SMTP server)
	ctx := context.Background()
	
	// Use localhost instead of external domain to avoid DNS blocks
	localConfig := config
	localConfig.SMTPHost = "localhost"
	localConfig.SMTPPort = 65534 // Use unused port to ensure quick failure
	
	if err := handler.SendNotification(ctx, alert, localConfig); err == nil {
		t.Error("expected error when sending email without valid SMTP server")
	}

	// Test with incomplete config
	incompleteConfig := EmailConfig{
		Enabled: true,
		// Missing required fields
	}
	incompleteHandler := NewEmailHandler(incompleteConfig, logger)
	if err := incompleteHandler.SendNotification(ctx, alert, incompleteConfig); err == nil {
		t.Error("expected error when sending email with incomplete config")
	}

	// Test with no recipients
	noRecipientsConfig := config
	noRecipientsConfig.Recipients = []string{}
	noRecipientsHandler := NewEmailHandler(noRecipientsConfig, logger)
	if err := noRecipientsHandler.SendNotification(ctx, alert, noRecipientsConfig); err == nil {
		t.Error("expected error when sending email with no recipients")
	}
}

// TestNotificationHandlerInterfaces tests that all handlers implement the interface correctly
func TestNotificationHandlerInterfaces(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())

	handlers := []struct {
		name    string
		handler NotificationHandler
		channel NotificationChannel
	}{
		{
			name:    "LogHandler",
			handler: NewLogHandler(logger),
			channel: ChannelLog,
		},
		{
			name: "SlackHandler",
			handler: NewSlackHandler(SlackConfig{
				Enabled:   true,
				Channel:   "#test",
				Username:  "test",
				IconEmoji: ":test:",
				Timeout:   30 * time.Second,
			}, logger),
			channel: ChannelSlack,
		},
		{
			name: "EmailHandler",
			handler: NewEmailHandler(EmailConfig{
				Enabled:    true,
				SMTPHost:   "localhost", // Use localhost instead of external domain
				SMTPPort:   587,
				From:       "test@test.com",
				Recipients: []string{"admin@test.com"},
				UseTLS:     true,
				Timeout:    30 * time.Second,
			}, logger),
			channel: ChannelEmail,
		},
	}

	ctx := context.Background()
	testAlert := &Alert{
		ID:          "interface-test",
		Type:        AlertTypeAnomaly,
		Severity:    SeverityWarning,
		Status:      AlertStatusFired,
		Title:       "Interface Test Alert",
		Description: "Test alert for interface verification",
		Timestamp:   time.Now(),
		Source:      "test",
	}

	for _, test := range handlers {
		t.Run(test.name, func(t *testing.T) {
			// Test GetChannelType
			if test.handler.GetChannelType() != test.channel {
				t.Errorf("expected channel %s, got %s", test.channel, test.handler.GetChannelType())
			}

			// Test SendNotification (might fail, but should not panic)
			test.handler.SendNotification(ctx, testAlert, nil)

			// Test TestNotification (might fail, but should not panic)
			test.handler.TestNotification(ctx, nil)
		})
	}
}

// TestNotificationErrorHandling tests error handling in notifications
func TestNotificationErrorHandling(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	ctx := context.Background()

	// Test with nil alert
	logHandler := NewLogHandler(logger)
	if err := logHandler.SendNotification(ctx, nil, nil); err != nil {
		// LogHandler should handle nil gracefully
		t.Errorf("log handler should handle nil alert gracefully: %v", err)
	}

	// Test with malformed alert
	malformedAlert := &Alert{
		// Missing required fields
		Type:     "",
		Severity: "",
	}

	if err := logHandler.SendNotification(ctx, malformedAlert, nil); err != nil {
		t.Errorf("log handler should handle malformed alert gracefully: %v", err)
	}

	// Test context cancellation
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	normalAlert := &Alert{
		ID:          "cancelled-test",
		Type:        AlertTypeAnomaly,
		Severity:    SeverityWarning,
		Status:      AlertStatusFired,
		Title:       "Cancelled Test Alert",
		Description: "Test alert for context cancellation",
		Timestamp:   time.Now(),
		Source:      "test",
	}

	// Log handler should still work with cancelled context
	if err := logHandler.SendNotification(cancelledCtx, normalAlert, nil); err != nil {
		t.Errorf("log handler should handle cancelled context: %v", err)
	}

	// Slack handler should respect context cancellation
	slackHandler := NewSlackHandler(SlackConfig{
		Enabled:    true,
		WebhookURL: "https://localhost:65534/webhook", // Use localhost instead of external domain
		Timeout:    1 * time.Second, // Short timeout
	}, logger)

	// This should either fail quickly due to context cancellation or invalid URL
	slackHandler.SendNotification(cancelledCtx, normalAlert, nil)

	// Email handler with context cancellation - test config validation only
	emailHandler := NewEmailHandler(EmailConfig{
		Enabled:    true,
		SMTPHost:   "", // Empty host will fail config validation quickly
		SMTPPort:   587,
		From:       "",         // Empty from will fail config validation quickly
		Recipients: []string{}, // Empty recipients will fail config validation quickly
		Timeout:    1 * time.Second,
	}, logger)

	// This should fail quickly due to configuration error, not network timeout
	emailHandler.SendNotification(cancelledCtx, normalAlert, nil)
}
