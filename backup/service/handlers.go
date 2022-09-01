package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/temporal-demo-apps/backup"
)

func GetStatus(w http.ResponseWriter, r *http.Request) {
	var status backup.Status
	status.Msg = "OK"
	status.Version = "1.0.0"

	json.NewEncoder(w).Encode(status)
}

func GetBackupState(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var backupId string = params["backupId"]
	backupState, _ := backupState.Load(backupId)

	json.NewEncoder(w).Encode(backupState)
}

func Quiesce(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var backupId string = params["backupId"]

	backupState.Store(backupId, "quiesced")

	result := ChaosMonkey("Quiesce")
	json.NewEncoder(w).Encode(result)
}

func UnQuiesce(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var backupId string = params["backupId"]

	backupState.Store(backupId, "unquiesced")

	result := ChaosMonkey("UnQuiesce")
	json.NewEncoder(w).Encode(result)
}

func Backup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var backupId string = params["backupId"]

	backupState.Store(backupId, "backup")

	result := ChaosMonkey("Backup")
	json.NewEncoder(w).Encode(result)
}

func ChaosMonkey(msg string) backup.Result {
	var result backup.Result

	if os.Getenv("ENABLE_CHAOS_MONKEY") == "true" {
		sleepTimer := rand.Intn(1250)
		time.Sleep(time.Duration(sleepTimer) * time.Millisecond)

		var code int
		rand.Seed(time.Now().UnixNano())
		if msg == "Backup" {
			min := 0
			max := 1
			code = rand.Intn(max-min+1) + min
		} else {
			min := 0
			max := 5
			code = rand.Intn(max-min+1) + min
		}

		result.Code = code
		result.Message = msg
	} else {
		result.Code = 0
		result.Message = msg
	}

	errorCode := strconv.Itoa(result.Code)
	fmt.Println("DEBUG: Message[" + result.Message + "] Code[" + errorCode + "]")

	return result
}
