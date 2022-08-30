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

func Quiesce(hostname, port string) (Result, error) {
	var result Result
	req, err := http.NewRequest("GET", "http://"+hostname+":"+port+"/quiesce", nil)

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

func Backup(hostname, port string) (Result, error) {
	var result Result
	req, err := http.NewRequest("GET", "http://"+hostname+":"+port+"/backup", nil)

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

func UnQuiesce(hostname, port string) (Result, error) {
	var result Result
	req, err := http.NewRequest("GET", "http://"+hostname+":"+port+"/unquiesce", nil)

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
