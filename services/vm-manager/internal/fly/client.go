package fly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL = "https://api.machines.dev/v1"
)

type Client struct {
	apiToken string
	appName  string
	client   *http.Client
}

type Machine struct {
	ID            string                 `json:"id"`
	AppName       string                 `json:"app"`
	Name          string                 `json:"name"`
	State         string                 `json:"state"` // "started", "stopped", "destroyed"
	Region        string                 `json:"region"`
	ImageRef      MachineImage           `json:"image"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
	Config        MachineConfig          `json:"config"`
	ProcessGroup  string                 `json:"process_group"`
	InstanceID    string                 `json:"instance_id"`
	PrivateIP     string                 `json:"private_ip"`
	Checks        []MachineCheck         `json:"checks"`
	Events        []interface{}          `json:"events"`
}

type MachineImage struct {
	Ref string `json:"ref"`
}

type MachineConfig struct {
	Image      string                 `json:"image"`
	Guest      GuestConfig            `json:"guest"`
	Env        map[string]string      `json:"env"`
	Mounts     []Mount                `json:"mounts"`
	Services   []Service              `json:"services"`
	Restart    RestartPolicy          `json:"restart"`
	AutoDestroy bool                  `json:"auto_destroy"`
}

type GuestConfig struct {
	CPUs      int `json:"cpus"`
	MemoryMB  int `json:"memory_mb"`
	GPUs      int `json:"gpus,omitempty"`
}

type Mount struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Path        string `json:"path"`
	ReadOnly    bool   `json:"read_only,omitempty"`
}

type Service struct {
	Protocol   string `json:"protocol"`
	InternalPort int `json:"internal_port"`
	Ports      []Port `json:"ports"`
	Checks     []ServiceCheck `json:"checks"`
}

type Port struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type ServiceCheck struct {
	Type     string `json:"type"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}

type RestartPolicy struct {
	Policy string `json:"policy"` // "no", "on-failure", "always"
}

type MachineCheck struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Output      string `json:"output"`
}

// CreateMachineInput for creating a new machine
type CreateMachineInput struct {
	Name   string         `json:"name"`
	Region string         `json:"region"`
	Config MachineConfig  `json:"config"`
}

// CreateMachineResponse for creation result
type CreateMachineResponse struct {
	Machine *Machine `json:"machine"`
}

// NewClient creates a Fly.io API client
func NewClient(apiToken, appName string) *Client {
	return &Client{
		apiToken: apiToken,
		appName:  appName,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateMachine creates a new machine
func (c *Client) CreateMachine(input *CreateMachineInput) (*Machine, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/apps/%s/machines", baseURL, c.appName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("fly api error: %d - %s", resp.StatusCode, string(respBody))
	}

	var createResp CreateMachineResponse
	if err := json.Unmarshal(respBody, &createResp); err != nil {
		return nil, err
	}

	return createResp.Machine, nil
}

// GetMachine retrieves a machine by ID
func (c *Client) GetMachine(machineID string) (*Machine, error) {
	url := fmt.Sprintf("%s/apps/%s/machines/%s", baseURL, c.appName, machineID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fly api error: %d - %s", resp.StatusCode, string(respBody))
	}

	var machine Machine
	if err := json.Unmarshal(respBody, &machine); err != nil {
		return nil, err
	}

	return &machine, nil
}

// ListMachines lists all machines for the app
func (c *Client) ListMachines() ([]*Machine, error) {
	url := fmt.Sprintf("%s/apps/%s/machines", baseURL, c.appName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fly api error: %d - %s", resp.StatusCode, string(respBody))
	}

	var machines []*Machine
	if err := json.Unmarshal(respBody, &machines); err != nil {
		return nil, err
	}

	return machines, nil
}

// StopMachine stops a machine
func (c *Client) StopMachine(machineID string) error {
	url := fmt.Sprintf("%s/apps/%s/machines/%s/stop", baseURL, c.appName, machineID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fly api error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DestroyMachine destroys a machine
func (c *Client) DestroyMachine(machineID string) error {
	url := fmt.Sprintf("%s/apps/%s/machines/%s", baseURL, c.appName, machineID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fly api error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// StartMachine starts a stopped machine
func (c *Client) StartMachine(machineID string) error {
	url := fmt.Sprintf("%s/apps/%s/machines/%s/start", baseURL, c.appName, machineID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fly api error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}
