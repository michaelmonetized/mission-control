package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/michaelmonetized/mission-control/services/vm-manager/internal/relay"
	"github.com/michaelmonetized/mission-control/services/vm-manager/internal/vm"
)

// RegisterRoutes registers all API endpoints
func RegisterRoutes(e *echo.Echo, vmManager *vm.Manager) {
	// Initialize terminal relay
	terminalRelay := relay.NewTerminalRelay(1000) // max 1000 concurrent connections

	// VM endpoints
	vmGroup := e.Group("/api/vms")
	vmGroup.POST("", handleCreateVM(vmManager))
	vmGroup.GET("/:vm_id", handleGetVM(vmManager))
	vmGroup.GET("/user/:user_id", handleListVMs(vmManager))
	vmGroup.POST("/:vm_id/stop", handleStopVM(vmManager))
	vmGroup.DELETE("/:vm_id", handleDestroyVM(vmManager))
	vmGroup.POST("/:vm_id/activity", handleUpdateActivity(vmManager))

	// Terminal WebSocket
	e.GET("/api/terminal/connect", func(c echo.Context) error {
		terminalRelay.HandleConnection(c.Response().Writer, c.Request())
		return nil
	})
	e.GET("/api/terminal/clients", handleListClients(terminalRelay))
	e.DELETE("/api/terminal/clients/:client_id", handleDisconnectClient(terminalRelay))

	// System endpoints
	e.GET("/api/system/stats", handleSystemStats(vmManager, terminalRelay))
	e.POST("/api/system/cleanup", handleCleanup(terminalRelay))
}

// handleCreateVM POST /api/vms
func handleCreateVM(vmManager *vm.Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		var input vm.CreateVMInput
		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		vmInstance, err := vmManager.CreateVM(c.Request().Context(), &input)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusCreated, vmInstance)
	}
}

// handleGetVM GET /api/vms/:vm_id
func handleGetVM(vmManager *vm.Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		vmID := c.Param("vm_id")
		vmInstance, err := vmManager.GetVM(vmID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, vmInstance)
	}
}

// handleListVMs GET /api/vms/user/:user_id
func handleListVMs(vmManager *vm.Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Param("user_id")
		vms := vmManager.ListVMs(userID)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user_id": userID,
			"vms":     vms,
			"count":   len(vms),
		})
	}
}

// handleStopVM POST /api/vms/:vm_id/stop
func handleStopVM(vmManager *vm.Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		vmID := c.Param("vm_id")
		if err := vmManager.StopVM(vmID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "stopped"})
	}
}

// handleDestroyVM DELETE /api/vms/:vm_id
func handleDestroyVM(vmManager *vm.Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		vmID := c.Param("vm_id")
		if err := vmManager.DestroyVM(vmID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "destroyed"})
	}
}

// handleUpdateActivity POST /api/vms/:vm_id/activity
func handleUpdateActivity(vmManager *vm.Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		vmID := c.Param("vm_id")
		if err := vmManager.UpdateActivity(vmID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"activity": "updated"})
	}
}

// handleListClients GET /api/terminal/clients
func handleListClients(tr *relay.TerminalRelay) echo.HandlerFunc {
	return func(c echo.Context) error {
		vmID := c.QueryParam("vm_id")
		var clients []*relay.TerminalClient
		if vmID != "" {
			clients = tr.ListClients(vmID)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"clients": clients,
			"count":   len(clients),
		})
	}
}

// handleDisconnectClient DELETE /api/terminal/clients/:client_id
func handleDisconnectClient(tr *relay.TerminalRelay) echo.HandlerFunc {
	return func(c echo.Context) error {
		clientID := c.Param("client_id")
		if err := tr.DisconnectClient(clientID); err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "disconnected"})
	}
}

// handleSystemStats GET /api/system/stats
func handleSystemStats(vmManager *vm.Manager, tr *relay.TerminalRelay) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"running_vms":       vmManager.RunningCount(),
			"connected_clients": tr.ClientCount(),
			"timestamp":         time.Now().Unix(),
		})
	}
}

// handleCleanup POST /api/system/cleanup
func handleCleanup(tr *relay.TerminalRelay) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Clean up stale connections (idle > 1 hour)
		tr.CleanupStaleConnections(3600 * time.Second)
		return c.JSON(http.StatusOK, map[string]string{"status": "cleanup_complete"})
	}
}
