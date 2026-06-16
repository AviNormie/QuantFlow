package monitoring

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/posthog/posthog-go"
)

// Clients holds initialized Sentry and PostHog clients for a service.
type Clients struct {
	ServiceName    string
	SentryEnabled  bool
	PostHogEnabled bool
	PostHog        posthog.Client
}

// Init configures Sentry and PostHog from environment variables.
// Missing DSN/API key disables that integration without failing startup.
func Init(serviceName string) (*Clients, error) {
	clients := &Clients{ServiceName: serviceName}

	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		sampleRate, err := parseSampleRate(os.Getenv("SENTRY_TRACES_SAMPLE_RATE"))
		if err != nil {
			return nil, fmt.Errorf("invalid SENTRY_TRACES_SAMPLE_RATE: %w", err)
		}

		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			Environment:      envOr("SENTRY_ENVIRONMENT", "development"),
			Release:          os.Getenv("SENTRY_RELEASE"),
			ServerName:       serviceName,
			EnableTracing:    true,
			TracesSampleRate: sampleRate,
		}); err != nil {
			return nil, fmt.Errorf("sentry init: %w", err)
		}

		clients.SentryEnabled = true
	}

	if apiKey := os.Getenv("POSTHOG_API_KEY"); apiKey != "" {
		client, err := posthog.NewWithConfig(apiKey, posthog.Config{
			Endpoint: envOr("POSTHOG_HOST", "https://us.i.posthog.com"),
		})
		if err != nil {
			return nil, fmt.Errorf("posthog init: %w", err)
		}

		clients.PostHog = client
		clients.PostHogEnabled = true
	}

	return clients, nil
}

// AttachGin registers monitoring middleware on a Gin engine.
func (c *Clients) AttachGin(r *gin.Engine) {
	if c.SentryEnabled {
		r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	}

	if c.PostHogEnabled {
		r.Use(c.postHogMiddleware())
	}
}

// CaptureError reports an error to Sentry and PostHog.
func (c *Clients) CaptureError(err error, tags map[string]string) {
	if err == nil {
		return
	}

	if c.SentryEnabled {
		hub := sentry.CurrentHub().Clone()
		for key, value := range tags {
			hub.Scope().SetTag(key, value)
		}
		hub.CaptureException(err)
	}

	if c.PostHogEnabled {
		props := posthog.NewProperties().
			Set("service", c.ServiceName).
			Set("error", err.Error())

		for key, value := range tags {
			props.Set(key, value)
		}

		_ = c.PostHog.Enqueue(posthog.Capture{
			DistinctId: c.ServiceName,
			Event:      "backend_error",
			Properties: props,
		})
	}
}

// CaptureEvent sends a custom PostHog event.
func (c *Clients) CaptureEvent(distinctID, event string, properties map[string]any) {
	if !c.PostHogEnabled || event == "" {
		return
	}

	if distinctID == "" {
		distinctID = c.ServiceName
	}

	props := posthog.NewProperties().Set("service", c.ServiceName)
	for key, value := range properties {
		props.Set(key, value)
	}

	_ = c.PostHog.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      event,
		Properties: props,
	})
}

// Close flushes Sentry and shuts down PostHog.
func (c *Clients) Close() {
	if c.SentryEnabled {
		sentry.Flush(2 * time.Second)
	}

	if c.PostHog != nil {
		_ = c.PostHog.Close()
	}
}

func (c *Clients) postHogMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/health" {
			ctx.Next()
			return
		}

		start := time.Now()
		ctx.Next()

		path := ctx.FullPath()
		if path == "" {
			path = ctx.Request.URL.Path
		}

		_ = c.PostHog.Enqueue(posthog.Capture{
			DistinctId: c.ServiceName,
			Event:      "api_request",
			Properties: posthog.NewProperties().
				Set("service", c.ServiceName).
				Set("method", ctx.Request.Method).
				Set("path", path).
				Set("status", ctx.Writer.Status()).
				Set("duration_ms", time.Since(start).Milliseconds()),
		})
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseSampleRate(raw string) (float64, error) {
	if raw == "" {
		return 0.2, nil
	}

	rate, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, err
	}
	if rate < 0 || rate > 1 {
		return 0, fmt.Errorf("must be between 0 and 1")
	}

	return rate, nil
}
