package middleware

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"xrf197ilz35aq/internal/random"
)

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer // // A buffer for the response body
}

// WriteHeader writes the status code to the response.
func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Header returns the headers of the underlying response writer.
func (w *responseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b) // Write to the buffer
	return w.ResponseWriter.Write(b)
}

// LoggerHandler is a middleware that logs requests.
type LoggerHandler struct {
	logger *slog.Logger
}

func (lh *LoggerHandler) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := GenerateRequestId()
		start := time.Now()
		loggerWithReqId := lh.logger.With("requestId", requestId)

		path := r.URL.Path
		loggerWithReqId.Info("event=incomingRequest", "method", r.Method, "url", path, "remoteAddr", r.RemoteAddr)

		// Wrap the response writer to capture the status code.
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		wrappedWriter.Header().Set("Request-Trace-Id", requestId)

		// Call the next handler.
		next.ServeHTTP(wrappedWriter, r)

		// Stop the timer.
		timeTaken := time.Since(start)
		status := wrappedWriter.status

		if status >= 400 {
			errBody := wrappedWriter.body.String()
			loggerWithReqId.Error("event=response", "url", path, "timeTaken", timeTaken, "error", errBody)
		} else {
			loggerWithReqId.Info("event=response", "url", path, "status")
		}
	})
}

func GenerateRequestId() string {
	uniqueStr, err := random.TimeBasedString(time.Now().Unix(), 21)
	if err != nil {
		return strconv.Itoa(int(random.PositiveInt64()))
	}

	uniqueInt64 := random.PositiveInt64()
	uniqueInt64Str := strconv.Itoa(int(uniqueInt64))

	if len(uniqueInt64Str) > 10 {
		uniqueInt64Str = uniqueInt64Str[2:]
	}

	partStr := uniqueStr[0:12]

	return fmt.Sprintf("%s.%s", uniqueInt64Str, partStr)
}
