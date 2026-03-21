package vm

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/michaelmonetized/mission-control/services/vm-manager/internal/fly"
	"github.com/michaelmonetized/mission-control/services/vm-manager/internal/metrics"
)

type Manager struct {
	flyClient    *fly.Client
	metricsTracker *metrics.Tracker
	maxVMs       int

	// Track VMs locally for fast access
	vms   map[string]*VMInstance
	vmsMu sync.RWMutex

	// Track user -> VM count for scaling limits
	userVMs   map[string][]string
	userVMMu  sync.RWMutex

	// Track org limits
	orgLimits map[string]int
	orgMu     sync.RWMutex
}

type VMInstance struct {
	ID           string
	UserID       string
	OrgID        string
	RepoURL      string
	CreatedAt    time.Time
	LastActivity time.Time
	Status       string // "starting", "running", "stopping", "stopped"
	MachineID    string // Fly machine ID
	TerminalURL  string // ws://machine-id.fly.dev/terminal
	CPUCount     int
	MemoryMB     int
	BilledMinutes float64
}

type CreateVMInput struct {
	UserID    string
	OrgID     string
	RepoURL   string
	RepoRef   string // branch or commit
	APIKey    string // Claude API key
	Region    string // Fly region
	CPUs      int
	MemoryMB  int
}

// NewManager creates a new VM manager
func NewManager(flyClient *fly.Client, metricsTracker *metrics.Tracker, maxVMs int) *Manager {
	return &Manager{
		flyClient:      flyClient,
		metricsTracker: metricsTracker,
		maxVMs:         maxVMs,
		vms:            make(map[string]*VMInstance),
		userVMs:        make(map[string][]string),
		orgLimits:      make(map[string]int),
	}
}

// CreateVM spins up a new VM with the given configuration
func (m *Manager) CreateVM(ctx context.Context, input *CreateVMInput) (*VMInstance, error) {
	// Check org limit
	m.orgMu.RLock()
	currentCount := m.orgLimits[input.OrgID]
	m.orgMu.RUnlock()

	if currentCount >= m.maxVMs {
		return nil, fmt.Errorf("org %s has reached max VMs (%d)", input.OrgID, m.maxVMs)
	}

	// Default region
	region := input.Region
	if region == "" {
		region = "ord" // Chicago
	}

	// Default resources
	cpus := input.CPUs
	if cpus == 0 {
		cpus = 2
	}
	memMB := input.MemoryMB
	if memMB == 0 {
		memMB = 4096
	}

	// Create Fly machine
	flyInput := &fly.CreateMachineInput{
		Name:   fmt.Sprintf("mc-%s-%d", input.UserID, time.Now().UnixNano()),
		Region: region,
		Config: fly.MachineConfig{
			Image: "ghcr.io/michaelmonetized/mc-workspace:latest",
			Guest: fly.GuestConfig{
				CPUs:     cpus,
				MemoryMB: memMB,
			},
			Env: map[string]string{
				"REPO_URL":          input.RepoURL,
				"REPO_REF":          input.RepoRef,
				"ANTHROPIC_API_KEY": input.APIKey,
				"MC_USER_ID":        input.UserID,
				"MC_ORG_ID":         input.OrgID,
			},
			Restart: fly.RestartPolicy{
				Policy: "no", // Don't auto-restart
			},
			AutoDestroy: true,
		},
	}

	machine, err := m.flyClient.CreateMachine(flyInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create fly machine: %w", err)
	}

	// Create local instance
	vmInstance := &VMInstance{
		ID:           fmt.Sprintf("%s-%s", input.OrgID, machine.ID),
		UserID:       input.UserID,
		OrgID:        input.OrgID,
		RepoURL:      input.RepoURL,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Status:       "starting",
		MachineID:    machine.ID,
		TerminalURL:  fmt.Sprintf("wss://%s.fly.dev/terminal", machine.ID),
		CPUCount:     cpus,
		MemoryMB:     memMB,
		BilledMinutes: 0,
	}

	// Store locally
	m.vms[vmInstance.ID] = vmInstance
	m.userVMs[input.UserID] = append(m.userVMs[input.UserID], vmInstance.ID)

	// Update org limit
	m.orgMu.Lock()
	m.orgLimits[input.OrgID]++
	m.orgMu.Unlock()

	log.Printf("✅ Created VM %s for user %s (machine: %s)", vmInstance.ID, input.UserID, machine.ID)
	m.metricsTracker.RecordVMCreated(input.OrgID, input.UserID)

	return vmInstance, nil
}

// GetVM retrieves a VM instance
func (m *Manager) GetVM(vmID string) (*VMInstance, error) {
	m.vmsMu.RLock()
	defer m.vmsMu.RUnlock()

	vm, exists := m.vms[vmID]
	if !exists {
		return nil, fmt.Errorf("vm %s not found", vmID)
	}

	return vm, nil
}

// ListVMs lists VMs for a user
func (m *Manager) ListVMs(userID string) []*VMInstance {
	m.userVMMu.RLock()
	vmIDs := m.userVMs[userID]
	m.userVMMu.RUnlock()

	m.vmsMu.RLock()
	defer m.vmsMu.RUnlock()

	var result []*VMInstance
	for _, vmID := range vmIDs {
		if vm, exists := m.vms[vmID]; exists {
			result = append(result, vm)
		}
	}

	return result
}

// UpdateActivity records user activity on a VM
func (m *Manager) UpdateActivity(vmID string) error {
	m.vmsMu.Lock()
	defer m.vmsMu.Unlock()

	vm, exists := m.vms[vmID]
	if !exists {
		return fmt.Errorf("vm %s not found", vmID)
	}

	vm.LastActivity = time.Now()
	return nil
}

// StopVM stops a VM
func (m *Manager) StopVM(vmID string) error {
	m.vmsMu.Lock()
	vm, exists := m.vms[vmID]
	if !exists {
		m.vmsMu.Unlock()
		return fmt.Errorf("vm %s not found", vmID)
	}
	m.vmsMu.Unlock()

	// Stop in Fly
	if err := m.flyClient.StopMachine(vm.MachineID); err != nil {
		return err
	}

	m.vmsMu.Lock()
	vm.Status = "stopped"
	m.vmsMu.Unlock()

	log.Printf("⏹️  Stopped VM %s", vmID)
	return nil
}

// DestroyVM destroys a VM
func (m *Manager) DestroyVM(vmID string) error {
	m.vmsMu.Lock()
	vm, exists := m.vms[vmID]
	if !exists {
		m.vmsMu.Unlock()
		return fmt.Errorf("vm %s not found", vmID)
	}
	m.vmsMu.Unlock()

	// Destroy in Fly
	if err := m.flyClient.DestroyMachine(vm.MachineID); err != nil {
		return err
	}

	// Calculate billed minutes
	billedMinutes := time.Since(vm.CreatedAt).Minutes()

	// Record usage
	m.metricsTracker.RecordUsage(vm.OrgID, vm.UserID, billedMinutes)

	// Remove from local storage
	m.vmsMu.Lock()
	delete(m.vms, vmID)
	m.vmsMu.Unlock()

	m.userVMMu.Lock()
	userVMs := m.userVMs[vm.UserID]
	for i, id := range userVMs {
		if id == vmID {
			m.userVMs[vm.UserID] = append(userVMs[:i], userVMs[i+1:]...)
			break
		}
	}
	m.userVMMu.Unlock()

	// Update org limit
	m.orgMu.Lock()
	m.orgLimits[vm.OrgID]--
	m.orgMu.Unlock()

	log.Printf("🗑️  Destroyed VM %s (%.2f minutes billed)", vmID, billedMinutes)
	return nil
}

// RunningCount returns the total number of running VMs
func (m *Manager) RunningCount() int {
	m.vmsMu.RLock()
	defer m.vmsMu.RUnlock()
	return len(m.vms)
}

// StartHealthCheck periodically checks VM health
func (m *Manager) StartHealthCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkHealth()
		}
	}
}

func (m *Manager) checkHealth() {
	m.vmsMu.RLock()
	vmIDs := make([]string, 0, len(m.vms))
	for id := range m.vms {
		vmIDs = append(vmIDs, id)
	}
	m.vmsMu.RUnlock()

	for _, vmID := range vmIDs {
		m.vmsMu.RLock()
		vm := m.vms[vmID]
		m.vmsMu.RUnlock()

		machine, err := m.flyClient.GetMachine(vm.MachineID)
		if err != nil {
			log.Printf("⚠️  Health check failed for %s: %v", vmID, err)
			continue
		}

		m.vmsMu.Lock()
		vm.Status = machine.State
		m.vmsMu.Unlock()
	}
}

// StartGracefulShutdown kills VMs idle longer than maxDuration
func (m *Manager) StartGracefulShutdown(ctx context.Context, maxIdleDuration time.Duration) {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.performGracefulShutdown(maxIdleDuration)
		}
	}
}

func (m *Manager) performGracefulShutdown(maxIdleDuration time.Duration) {
	m.vmsMu.RLock()
	vmIDs := make([]string, 0, len(m.vms))
	for id := range m.vms {
		vmIDs = append(vmIDs, id)
	}
	m.vmsMu.RUnlock()

	now := time.Now()
	for _, vmID := range vmIDs {
		m.vmsMu.RLock()
		vm := m.vms[vmID]
		m.vmsMu.RUnlock()

		idleDuration := now.Sub(vm.LastActivity)
		if idleDuration > maxIdleDuration {
			log.Printf("⏱️  VM %s idle for %.0f hours, destroying", vmID, idleDuration.Hours())
			_ = m.DestroyVM(vmID)
		}
	}
}
