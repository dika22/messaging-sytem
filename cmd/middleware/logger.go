package middleware

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"multi-tenant-service/metrics"
)

func LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c) // lanjut ke handler
		stop := time.Now()
		// Logging info request
		log.Info().
			Str("method", c.Request().Method).
			Str("uri", c.Request().RequestURI).
			Str("remote_ip", c.RealIP()).
			Int("status", c.Response().Status).
			Dur("latency", stop.Sub(start)).
			Msg("Handled request")

		return err
	}
}

func MonitoringMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		duration := time.Since(start)

		path := c.Path() // raw route (e.g., "/hello")
		method := c.Request().Method
		status := c.Response().Status

		metrics.HttpRequestsTotal.WithLabelValues(method, path, formatStatus(status)).Inc()
		metrics.HttpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())

		return err
	}
}

func formatStatus(code int) string {
	return fmt.Sprintf("%d", code)
}
