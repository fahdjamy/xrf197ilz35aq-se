package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type RequestHandler interface {
	RegisterRoutes(serveMux *http.ServeMux)
}

func logLatency(startTime time.Time, event string, logger slog.Logger) {
	msg := fmt.Sprintf("event=%s", event)
	logger.Info(msg, "latency", time.Since(startTime))
}
