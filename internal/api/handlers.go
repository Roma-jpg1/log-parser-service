package api

import (
	"awesomeProject/internal/db"
	"awesomeProject/internal/parser"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type Handler struct {
	repo *db.Repository
}

func NewHandler(repo *db.Repository) *Handler {
	return &Handler{repo: repo}
}

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

func (h *Handler) ParseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	var req ParseRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.CSVPath == "" {
		WriteError(w, http.StatusBadRequest, "CSV path is required")
		return
	}

	if req.SharpPath == "" {
		WriteError(w, http.StatusBadRequest, "Sharp path is required")
		return
	}

	if !isDataPath(req.CSVPath) || !isDataPath(req.SharpPath) {
		WriteError(w, http.StatusBadRequest, "log paths must be inside data/")
		return
	}

	logID, err := h.repo.CreateLog(req.CSVPath)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create log")
		return
	}

	parsed, err := parser.ParseFiles(req.CSVPath, req.SharpPath)
	if err != nil {
		_ = h.repo.MarkLogFailed(logID, err.Error())
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.repo.SaveParsedLog(logID, parsed); err != nil {
		_ = h.repo.MarkLogFailed(logID, err.Error())
		WriteError(w, http.StatusInternalServerError, "Failed to save parsed log")
		return
	}

	if err := h.repo.MarkLogSuccess(logID, len(parsed.Nodes), len(parsed.Ports)); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to update log status")
		return
	}

	WriteJson(w, http.StatusOK, ParseResponse{
		LogID: logID,
		Msg:   "log parsed",
	})
}

func (h *Handler) TopologyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	logID, err := extractID(r.URL.Path, "/api/v1/topology/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid log ID")
		return
	}

	nodes, err := h.repo.GetNodesByLogID(logID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get topology")
		return
	}

	groups := map[string][]any{
		"hosts":    {},
		"switches": {},
	}
	for _, node := range nodes {
		switch node.Type {
		case "host":
			groups["hosts"] = append(groups["hosts"], node)
		case "switch":
			groups["switches"] = append(groups["switches"], node)
		}
	}

	WriteJson(w, http.StatusOK, map[string]interface{}{
		"log_id": logID,
		"nodes":  nodes,
		"groups": groups,
	})
}

func (h *Handler) NodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	nodeID, err := extractID(r.URL.Path, "/api/v1/node/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid node ID")
		return
	}

	node, err := h.repo.GetNodeByID(nodeID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Node not found")
		return
	}

	infos, err := h.repo.GetNodeInfoByNodeID(nodeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get node info")
		return
	}

	WriteJson(w, http.StatusOK, map[string]interface{}{
		"node": node,
		"info": infos,
	})

}

func (h *Handler) PortHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	nodeID, err := extractID(r.URL.Path, "/api/v1/port/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid node ID")
		return
	}

	ports, err := h.repo.GetPortsByNodeID(nodeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get ports")
		return
	}

	WriteJson(w, http.StatusOK, map[string]interface{}{
		"node_id": nodeID,
		"ports":   ports,
	})
}

func (h *Handler) LogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	logID, err := extractID(r.URL.Path, "/api/v1/log/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid log ID")
		return
	}

	item, err := h.repo.GetLogByID(logID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Log not found")
		return
	}

	WriteJson(w, http.StatusOK, item)
}

func WriteJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
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

func isDataPath(path string) bool {
	cleanPath := filepath.ToSlash(filepath.Clean(path))
	return cleanPath == "data" || strings.HasPrefix(cleanPath, "data/")
}
