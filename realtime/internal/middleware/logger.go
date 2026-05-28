package middleware

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

const (
	xRealIPHeader          = "X-Real-IP"
	xForwardedForHeader    = "X-Forwarded-For"
	requestBodyLengthLimit = 1 * 256 // 256 байт
)

var (
	ErrInteralServerError = errors.New("something went wrong")
)

type LoggerResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *LoggerResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *LoggerResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

type LoggerLimitWriter struct {
	Destination io.Writer
	Remaining   int
}

func (w *LoggerLimitWriter) Write(p []byte) (int, error) {
	if w.Remaining <= 0 {
		return len(p), nil
	}

	toWrite := min(len(p), w.Remaining)
	_, err := w.Destination.Write(p[:toWrite])
	if err != nil {
		return 0, fmt.Errorf("destination.Write: %w", err)
	}

	w.Remaining -= toWrite

	return toWrite, nil
}

func NewLoggerLimitWriter(destination io.Writer, limit int) *LoggerLimitWriter {
	return &LoggerLimitWriter{
		Destination: destination,
		Remaining:   limit,
	}
}

type LoggerBufferedReader struct {
	io.Reader
	io.Closer
}

func NewLoggerBufferedReader(source io.ReadCloser, buf io.Writer) *LoggerBufferedReader {
	return &LoggerBufferedReader{
		Reader: io.TeeReader(source, buf),
		Closer: source,
	}
}

func GetRealIP(r *http.Request) string {
	if ip := r.Header.Get(xRealIPHeader); ip != "" {
		return ip
	}

	if ip := r.Header.Get(xForwardedForHeader); ip != "" {
		return strings.Split(ip, ",")[0]
	}

	return r.RemoteAddr
}

func ShouldLogRequestBody(code int) bool {
	return (code >= 400) && (code < 600)
}

func LoggerMiddleware(logger *zerolog.Logger) mux.MiddlewareFunc {
	buffersPool := sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestStart := time.Now()

			requestLogger := logger.With().Str("request-id", uuid.New().String()).Logger()

			buf := buffersPool.Get().(*bytes.Buffer)
			defer func() {
				buf.Reset()
				buffersPool.Put(buf)
			}()

			limitWriter := NewLoggerLimitWriter(buf, requestBodyLengthLimit)
			r.Body = NewLoggerBufferedReader(r.Body, limitWriter)

			ctx := requestLogger.WithContext(r.Context())

			responseWriter := &LoggerResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			next.ServeHTTP(responseWriter, r.WithContext(ctx))

			requestDuration := time.Since(requestStart)

			var logEvent *zerolog.Event
			isResponseError := ShouldLogRequestBody(responseWriter.statusCode)
			if isResponseError {
				logEvent = requestLogger.Error()
			} else {
				logEvent = requestLogger.Info()
			}

			logEvent = logEvent.
				Str("method", r.Method).
				Str("remote_addr", r.RemoteAddr).
				Str("url", r.URL.Path).
				Dur("work_time", requestDuration).
				Int("status", responseWriter.statusCode).
				Str("user_agent", r.UserAgent()).
				Str("host", r.Host).
				Str("real_ip", GetRealIP(r)).
				Int64("content_length", r.ContentLength).
				Str("start_time", requestStart.Format(time.RFC3339)).
				Str("duration_human", requestDuration.String()).
				Int64("duration_ms", requestDuration.Milliseconds())

			if isResponseError {
				if r.Body != nil {
					remaning := io.LimitReader(r.Body, requestBodyLengthLimit)
					_, _ = io.Copy(io.Discard, remaning)
				}

				logEvent = logEvent.Bytes("body", buf.Bytes())
			}

			logEvent.Msg("request processed")
		})
	}
}
