package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/michaelmonetized/mission-control/services/vm-manager/internal/fly"
	"github.com/michaelmonetized/mission-control/services/vm-manager/internal/metrics"
	"github.com/michaelmonetized/mission-control/services/vm-manager/internal/vm"
)

func TestHealthEndpoint(t *testing.T) {
	e := echo.New()

	// Create a test server
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Create test VM manager (with mock Fly client)
	mockFly := &fly.Client{}
	metricsTracker := metrics.NewTracker()
	vmManager := vm.NewManager(mockFly, metricsTracker, 100)

	c := e.NewContext(req, rec)

	// Test health endpoint
	handler := func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status":     "ok",
			"runningVMs": vmManager.RunningCount(),
			"totalBilled": metricsTracker.TotalBilled(),
		})
	}

	if err := handler(c); err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}
}

func TestMetricsEndpoint(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	metricsTracker := metrics.NewTracker()

	// Record some test data
	metricsTracker.RecordVMCreated("org-123", "user-456")
	metricsTracker.RecordUsage("org-123", "user-456", 60.5)

	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, metricsTracker.Report())
	}

	if err := handler(c); err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var report map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if report["total_vms_created"] != float64(1) {
		t.Errorf("Expected 1 VM created, got %v", report["total_vms_created"])
	}
}

func TestCreateVMRequest(t *testing.T) {
	e := echo.New()

	payload := map[string]interface{}{
		"user_id":  "user-123",
		"org_id":   "org-456",
		"repo_url": "https://github.com/test/repo.git",
		"api_key":  "sk-test",
		"region":   "ord",
		"cpus":     2,
		"memory_mb": 4096,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/vms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	// Verify we can parse the request
	var input map[string]interface{}
	if err := c.Bind(&input); err != nil {
		t.Fatalf("Failed to bind request: %v", err)
	}

	if input["user_id"] != "user-123" {
		t.Errorf("Expected user_id 'user-123', got %v", input["user_id"])
	}
}

func TestVMManagerRunningCount(t *testing.T) {
	mockFly := &fly.Client{}
	metricsTracker := metrics.NewTracker()
	vmManager := vm.NewManager(mockFly, metricsTracker, 100)

	if vmManager.RunningCount() != 0 {
		t.Errorf("Expected 0 VMs initially, got %d", vmManager.RunningCount())
	}
}

func TestMetricsTrackerBilling(t *testing.T) {
	tracker := metrics.NewTracker()

	// Record 100 minutes of usage at $0.02/min = $2.00
	tracker.RecordUsage("org-123", "user-456", 100.0)

	totalBilled := tracker.TotalBilled()
	expected := 2.00

	if totalBilled != expected {
		t.Errorf("Expected $%.2f billed, got $%.2f", expected, totalBilled)
	}
}

func BenchmarkMetricsTracking(b *testing.B) {
	tracker := metrics.NewTracker()

	for i := 0; i < b.N; i++ {
		tracker.RecordVMCreated("org-123", "user-456")
		tracker.RecordUsage("org-123", "user-456", 60.5)
	}
}
