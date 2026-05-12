package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ParseResponse struct {
	LogID int    `json:"log_id"`
	Msg   string `json:"msg"`
}

type ParseRequest struct {
	CSVPath   string `json:"csv_path"`
	SharpPath string `json:"sharp_path"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func ParseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req ParseRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteError(w, http.StatusBadRequest, "Invalid request payload")
	}

	if req.CSVPath == "" {
		WriteError(w, http.StatusBadRequest, "CSV path is required")
	}

	if req.SharpPath == "" {
		WriteError(w, http.StatusBadRequest, "Sharp path is required")
	}

	WriteJson(w, http.StatusOK, ParseResponse{
		LogID: config.LogID,
		Msg:   req.CSVPath,
	})
}

func TopologyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	LogID, err := extractID(r.URL.Path, "api/v1/topology/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid log ID")
	}

	WriteJson(w, http.StatusOK, map[string]interface{}{
		"log_id":   LogID,
		"nodes":    []any{},
		"groups":   []any{},
		"messages": "topology handl",
	})
}

func NodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	nodeID, err := extractID(r.URL.Path, "/api/v1/node/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid node ID")
	}

	WriteJson(w, http.StatusOK, map[string]interface{}{
		"node_id":  nodeID,
		"messages": "node handl",
	})

}

func PortHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	nodeID, err := extractID(r.URL.Path, "/api/v1/port/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid node ID")
		return
	}

	WriteJson(w, http.StatusOK, map[string]interface{}{
		"node_id":  nodeID,
		"ports":    []any{},
		"messages": "port handl",
	})
}

func LogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	logID, err := extractID(r.URL.Path, "/api/v1/log/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid log ID")
		return
	}
	WriteJson(w, http.StatusOK, map[string]interface{}{
		"log_id":      logID,
		"status":      "stub",
		"nodes_count": 0,
		"ports_count": 0,
		"messages":    "log handl",
	})
}

func WriteJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json/")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("json encode error: %v\n", err)
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJson(w, status, ErrorResponse{Error: message})
}

func extractID(path, prefix string) (int, error) {
	rawID := strings.TrimPrefix(path, prefix)
	if rawID == "" {
		return 0, fmt.Errorf("invalid path: %s", path)
	}
	id, err := strconv.Atoi(rawID)
	if err != nil {
		return 0, fmt.Errorf("invalid path: %s", path)
	}
	return id, nil
}
