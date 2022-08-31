package backup

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type WorkflowResult struct {
	Code     int      `json:"code"`
	Messages []string `json:"message"`
}

func Quiesce(hostname, port, backupId string) (Result, error) {
	var result Result
	req, err := http.NewRequest("POST", "http://"+hostname+":"+port+"/quiesce/"+backupId, nil)

	if err != nil {
		return result, err
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return result, err
		}
	} else {
		return result, errors.New("Http Status Error [" + resp.Status + "]")
	}

	return result, nil
}

func Backup(hostname, port, backupId string) (Result, error) {
	var result Result
	req, err := http.NewRequest("POST", "http://"+hostname+":"+port+"/backup/"+backupId, nil)

	if err != nil {
		return result, err
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return result, err
		}
	} else {
		return result, errors.New("Http Status Error [" + resp.Status + "]")
	}

	return result, nil
}

func UnQuiesce(hostname, port, backupId string) (Result, error) {
	var result Result
	req, err := http.NewRequest("POST", "http://"+hostname+":"+port+"/unquiesce/"+backupId, nil)

	if err != nil {
		return result, err
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return result, err
		}
	} else {
		return result, errors.New("Http Status Error [" + resp.Status + "]")
	}

	return result, nil
}

func GetBackupState(hostname, port, backupId string) (string, error) {
	var backupState string
	req, err := http.NewRequest("POST", "http://"+hostname+":"+port+"/getBackupState/"+backupId, nil)

	if err != nil {
		return backupState, err
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return backupState, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(&backupState); err != nil {
			return backupState, err
		}
	} else {
		return backupState, errors.New("Http Status Error [" + resp.Status + "]")
	}

	return backupState, nil
}
