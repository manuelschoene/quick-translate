package system

import (
	"context"
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

// Sends desktop notifications. Wraps the Wails notification service, so none of its methods are bound to the frontend.
type Notifier struct {
	service *notifications.NotificationService
	files   *Files
	logger  *slog.Logger
	ready   bool // Set once the inner service has started; nothing is sent without it
}

// Creates the notifier. The connection to the desktop is made on startup.
func NewNotifier(files *Files, logger *slog.Logger) *Notifier {
	return &Notifier{
		service: notifications.New(),
		files:   files,
		logger:  logger.With("category", "notifier"),
	}
}

// Creates the categories file with narrow permissions before Wails does, then starts the inner service. Never fails, as missing notifications do not make the application unusable.
func (n *Notifier) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	if _, err := n.files.EnsureFile(NotificationCategories, []byte("{}")); err != nil {
		n.logger.Warn("create notification categories", slog.Any("error", err))
	}

	if err := n.service.ServiceStartup(ctx, options); err != nil {
		n.logger.Warn("connect to the notification service", slog.Any("error", err))
		return nil
	}

	n.ready = true

	return nil
}

// Stops the inner service, if it was started.
func (n *Notifier) ServiceShutdown() error {
	if !n.ready {
		return nil
	}

	return n.service.ServiceShutdown()
}

// Tells the user that the application could not start, with the error as the body. Does nothing without a connection to the desktop; the error is logged by the caller of the application run anyway.
//
//wails:ignore
func (n *Notifier) StartupFailed(err error) {
	if !n.ready {
		return
	}

	options := notifications.NotificationOptions{
		ID:    "startup-failure",
		Title: AppName + " could not start",
		Body:  err.Error(),
	}

	if err := n.service.SendNotification(options); err != nil {
		n.logger.Warn("send notification", slog.Any("error", err))
	}
}
