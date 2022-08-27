package main

import (
	"encoding/json"
	"net/http"
)

type Status struct {
	Msg     string `json:"msg"`
	Version string `json:"version"`
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	var status Status
	status.Msg = "OK"
	status.Version = "1.0.0"

	json.NewEncoder(w).Encode(status)
}
