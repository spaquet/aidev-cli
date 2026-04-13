package views

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
)

// NotificationType represents the severity of a notification
type NotificationType int

const (
	NotificationInfo NotificationType = iota
	NotificationSuccess
	NotificationWarning
	NotificationError
)

// Notification represents a toast message
type Notification struct {
	Type      NotificationType
	Message   string
	CreatedAt time.Time
	Duration  time.Duration // 0 = forever
}

// NotificationManager manages multiple notifications
type NotificationManager struct {
	notifications []Notification
	maxNotifs     int
}

// NewNotificationManager creates a new manager
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		notifications: []Notification{},
		maxNotifs:     5,
	}
}

// Add adds a new notification
func (nm *NotificationManager) Add(notifType NotificationType, message string, duration time.Duration) {
	notif := Notification{
		Type:      notifType,
		Message:   message,
		CreatedAt: time.Now(),
		Duration:  duration,
	}

	nm.notifications = append(nm.notifications, notif)

	// Keep only last N notifications
	if len(nm.notifications) > nm.maxNotifs {
		nm.notifications = nm.notifications[len(nm.notifications)-nm.maxNotifs:]
	}
}

// Cleanup removes expired notifications
func (nm *NotificationManager) Cleanup() {
	now := time.Now()
	filtered := []Notification{}

	for _, notif := range nm.notifications {
		// Keep notifications with no duration (forever)
		if notif.Duration == 0 {
			filtered = append(filtered, notif)
			continue
		}

		// Keep notifications within duration
		if now.Sub(notif.CreatedAt) < notif.Duration {
			filtered = append(filtered, notif)
		}
	}

	nm.notifications = filtered
}

// Render renders all notifications as a string
func (nm *NotificationManager) Render(width int) string {
	if len(nm.notifications) == 0 {
		return ""
	}

	var output string

	for _, notif := range nm.notifications {
		line := nm.renderNotification(notif)
		output += line + "\n"
	}

	return output
}

func (nm *NotificationManager) renderNotification(notif Notification) string {
	var prefix string
	var style lipgloss.Style

	switch notif.Type {
	case NotificationSuccess:
		prefix = "[OK]"
		style = StyleSuccess
	case NotificationWarning:
		prefix = "[!]"
		style = StyleWarning
	case NotificationError:
		prefix = "[ERROR]"
		style = StyleError
	default:
		prefix = "[i]"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("#87CEEB"))
	}

	return style.Render(fmt.Sprintf("%s %s", prefix, notif.Message))
}

// Count returns the number of active notifications
func (nm *NotificationManager) Count() int {
	return len(nm.notifications)
}

// Clear removes all notifications
func (nm *NotificationManager) Clear() {
	nm.notifications = []Notification{}
}
