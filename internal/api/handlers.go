package api

import (
	"fmt"
	"net/http"
)

type ParseRequest struct {
	LogID int    `json:"log_id"`
	Msg   string `json:"msg"`
}

type ParseResponse struct {
	CSVPath   string `json:"csv_path"`
	SharpPath string `json:"sharp_path"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
