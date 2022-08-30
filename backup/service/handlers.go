package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

type Status struct {
	Msg     string `json:"msg"`
	Version string `json:"version"`
}

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	var status Status
	status.Msg = "OK"
	status.Version = "1.0.0"

	json.NewEncoder(w).Encode(status)
}

func Quiesce(w http.ResponseWriter, r *http.Request) {
	result := ChaosMonkey("Quiesce")
	json.NewEncoder(w).Encode(result)
}

func UnQuiesce(w http.ResponseWriter, r *http.Request) {
	result := ChaosMonkey("UnQuiesce")
	json.NewEncoder(w).Encode(result)
}

func Backup(w http.ResponseWriter, r *http.Request) {
	result := ChaosMonkey("Backup")
	json.NewEncoder(w).Encode(result)
}

func ChaosMonkey(msg string) Result {
	var result Result

	sleepTimer := rand.Intn(15)
	time.Sleep(time.Duration(sleepTimer) * time.Second)

	code := rand.Intn(5)

	result.Code = code
	result.Message = msg

	return result
}
