package observability

import (
	"context"
	"io"
	"log/slog"
	"strings"

	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	slogmulti "github.com/samber/slog-multi"
)

func Redact(groups []string, a slog.Attr) slog.Attr {
	key := strings.ToLower(strings.Join(groups, ".") + "." + a.Key)
	for _, sensitive := range []string{"password", "token", "secret", "cookie", "authorization", "credential", "dsn", "database_url"} {
		if strings.Contains(key, sensitive) {
			return slog.String(a.Key, "[REDACTED]")
		}
	}
	return a
}

func NewLogger(environment string, output io.Writer, hub *sentry.Hub) *slog.Logger {
	level := slog.LevelInfo
	if environment == "development" {
		level = slog.LevelDebug
	}
	console := slog.NewJSONHandler(output, &slog.HandlerOptions{Level: level, ReplaceAttr: Redact})
	if hub == nil {
		return slog.New(console)
	}
	remote := sentryslog.Option{
		Hub:         hub,
		EventLevel:  []slog.Level{slog.LevelError},
		LogLevel:    []slog.Level{},
		ReplaceAttr: Redact,
		AddSource:   true,
	}.NewSentryHandler(context.Background())
	return slog.New(slogmulti.Fanout(console, remote))
}

// ScrubEvent prevents HTTP credentials and payloads from entering crash reports.
func ScrubEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	if event.Request != nil {
		event.Request.Headers = nil
		event.Request.Cookies = ""
		event.Request.Data = ""
		event.Request.QueryString = ""
	}
	event.User = sentry.User{}
	return event
}
