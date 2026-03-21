package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type FlyIOClient struct {
	baseURL    string
	apiToken   string
	appName    string
	httpClient *http.Client
}

func NewFlyIOClient() *FlyIOClient {
	return &FlyIOClient{
		baseURL:    "https://api.machines.dev/v1",
		apiToken:   os.Getenv("FLY_API_TOKEN"),
		appName:    os.Getenv("FLY_APP_NAME"),
		httpClient: &http.Client{},
	}
}

// CreateMachine launches a new Fly Machine for a user's workspace
func (c *FlyIOClient) CreateMachine(userID, repoName, cloneURL string) (string, error) {
	req := map[string]interface{}{
		"name": fmt.Sprintf("%s-%s", userID, repoName),
		"config": map[string]interface{}{
			"image": "ghcr.io/michaelmonetized/mission-control-vm:latest",
			"env": map[string]string{
				"REPO_URL":     cloneURL,
				"USER_ID":      userID,
				"REPO_NAME":    repoName,
				"CLAUDE_TOKEN": os.Getenv("CLAUDE_API_TOKEN"), // TODO: per-user injection
			},
			"services": []map[string]interface{}{
				{
					"protocol": "tcp",
					"ports": []map[string]interface{}{
						{"port": 3000},
					},
				},
			},
		},
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/apps/%s/machines", c.baseURL, c.appName), bytes.NewReader(body))
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create machine: %s", string(bodyBytes))
	}

	var result struct {
		Machine struct {
			ID string `json:"id"`
		} `json:"machine"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Machine.ID, nil
}

// DeleteMachine stops and removes a Fly Machine
func (c *FlyIOClient) DeleteMachine(machineID string) error {
	httpReq, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/apps/%s/machines/%s", c.baseURL, c.appName, machineID), nil)
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete machine: %s", string(bodyBytes))
	}

	return nil
}

// GetMachineStatus fetches current status of a machine
func (c *FlyIOClient) GetMachineStatus(machineID string) (string, error) {
	httpReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/apps/%s/machines/%s", c.baseURL, c.appName, machineID), nil)
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Machine struct {
			State string `json:"state"`
		} `json:"machine"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Machine.State, nil
}
