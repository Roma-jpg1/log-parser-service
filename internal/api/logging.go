package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(data)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}

		logJSON(map[string]any{
			"level":       levelByStatus(status),
			"event":       "request",
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      status,
			"duration_ms": time.Since(started).Milliseconds(),
		})
	})
}

func logJSON(fields map[string]any) {
	payload, err := json.Marshal(fields)
	if err != nil {
		fmt.Printf(`{"level":"error","event":"log_encode","error":%q}`+"\n", err.Error())
		return
	}

	fmt.Println(string(payload))
}

func levelByStatus(status int) string {
	if status >= http.StatusInternalServerError {
		return "error"
	}
	if status >= http.StatusBadRequest {
		return "warn"
	}
	return "info"
}
