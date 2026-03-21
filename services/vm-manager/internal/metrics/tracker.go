package metrics

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

const (
	costPerMinute = 0.02 // $0.02 per minute
)

type Tracker struct {
	mu sync.RWMutex

	// Billing data
	usage     map[string]*UsageData // by org ID
	totalBilled float64

	// VM creation stats
	vmCreatedCount   int
	vmDestroyedCount int

	// Metrics path
	metricsPath string
}

type UsageData struct {
	OrgID    string
	UserID   string
	Minutes  float64
	Cost     float64
	Date     string
}

type Report struct {
	TotalVMsCreated   int               `json:"total_vms_created"`
	TotalVMsDestroyed int               `json:"total_vms_destroyed"`
	TotalBilled       float64           `json:"total_billed"`
	Usage             map[string]interface{} `json:"usage_by_org"`
	LastUpdated       int64             `json:"last_updated"`
}

func NewTracker() *Tracker {
	return &Tracker{
		usage:       make(map[string]*UsageData),
		metricsPath: "/tmp/vm-manager-metrics.json",
	}
}

func (t *Tracker) RecordVMCreated(orgID, userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.vmCreatedCount++
}

func (t *Tracker) RecordUsage(orgID, userID string, minutes float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	cost := minutes * costPerMinute
	t.totalBilled += cost

	key := orgID + "-" + userID
	t.usage[key] = &UsageData{
		OrgID:   orgID,
		UserID:  userID,
		Minutes: minutes,
		Cost:    cost,
		Date:    time.Now().Format("2006-01-02"),
	}
}

func (t *Tracker) RecordVMDestroyed(orgID, userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.vmDestroyedCount++
}

func (t *Tracker) TotalBilled() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.totalBilled
}

func (t *Tracker) Report() *Report {
	t.mu.RLock()
	defer t.mu.RUnlock()

	usageByOrg := make(map[string]interface{})
	for _, usage := range t.usage {
		if _, exists := usageByOrg[usage.OrgID]; !exists {
			usageByOrg[usage.OrgID] = []interface{}{}
		}
		usageByOrg[usage.OrgID] = append(usageByOrg[usage.OrgID].([]interface{}), map[string]interface{}{
			"user_id": usage.UserID,
			"minutes": usage.Minutes,
			"cost":    usage.Cost,
			"date":    usage.Date,
		})
	}

	return &Report{
		TotalVMsCreated:   t.vmCreatedCount,
		TotalVMsDestroyed: t.vmDestroyedCount,
		TotalBilled:       t.totalBilled,
		Usage:             usageByOrg,
		LastUpdated:       time.Now().Unix(),
	}
}

// StartPersistence periodically writes metrics to disk
func (t *Tracker) StartPersistence(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Save()
			return
		case <-ticker.C:
			t.Save()
		}
	}
}

// Save writes metrics to disk
func (t *Tracker) Save() error {
	report := t.Report()
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Printf("❌ Failed to marshal metrics: %v", err)
		return err
	}

	if err := os.WriteFile(t.metricsPath, data, 0644); err != nil {
		log.Printf("❌ Failed to write metrics: %v", err)
		return err
	}

	log.Printf("💾 Metrics saved: %s", t.metricsPath)
	return nil
}

// Load reads metrics from disk
func (t *Tracker) Load() error {
	data, err := os.ReadFile(t.metricsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet
		}
		return err
	}

	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.vmCreatedCount = report.TotalVMsCreated
	t.vmDestroyedCount = report.TotalVMsDestroyed
	t.totalBilled = report.TotalBilled

	return nil
}
