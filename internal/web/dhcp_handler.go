package web

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"pihole-analyzer/internal/dhcp"
	"pihole-analyzer/internal/types"
)

// DHCPHandler handles DHCP-related web requests
type DHCPHandler struct {
	dhcpServer dhcp.DHCPServer
	logger     *slog.Logger
}

// NewDHCPHandler creates a new DHCP web handler
func NewDHCPHandler(dhcpServer dhcp.DHCPServer, logger *slog.Logger) *DHCPHandler {
	return &DHCPHandler{
		dhcpServer: dhcpServer,
		logger:     logger,
	}
}

// DHCPStatusResponse represents the response for DHCP status
type DHCPStatusResponse struct {
	Status     *types.DHCPServerStatus `json:"status"`
	Statistics *types.DHCPStatistics   `json:"statistics"`
	Timestamp  string                  `json:"timestamp"`
}

// DHCPLeasesResponse represents the response for DHCP leases
type DHCPLeasesResponse struct {
	Leases    []types.DHCPLease `json:"leases"`
	Total     int               `json:"total"`
	Active    int               `json:"active"`
	Expired   int               `json:"expired"`
	Timestamp string            `json:"timestamp"`
}

// DHCPReservationsResponse represents the response for DHCP reservations
type DHCPReservationsResponse struct {
	Reservations []types.DHCPReservation `json:"reservations"`
	Total        int                     `json:"total"`
	Enabled      int                     `json:"enabled"`
	Timestamp    string                  `json:"timestamp"`
}

// RegisterDHCPRoutes registers DHCP-related routes with the HTTP server
func (s *Server) RegisterDHCPRoutes(dhcpServer dhcp.DHCPServer) {
	if dhcpServer == nil {
		s.logger.Warn("DHCP server is nil, skipping DHCP route registration")
		return
	}
	
	handler := NewDHCPHandler(dhcpServer, s.logger.GetSlogger())
	
	// Register DHCP API routes
	http.HandleFunc("/api/dhcp/status", handler.HandleStatus)
	http.HandleFunc("/api/dhcp/leases", handler.HandleLeases)
	http.HandleFunc("/api/dhcp/reservations", handler.HandleReservations)
	http.HandleFunc("/api/dhcp/lease/", handler.HandleLeaseAction)
	http.HandleFunc("/api/dhcp/reservation/", handler.HandleReservationAction)
	
	// Register DHCP web interface routes
	http.HandleFunc("/dhcp", handler.HandleDHCPPage)
	http.HandleFunc("/dhcp/", handler.HandleDHCPPage)
	
	s.logger.Info("DHCP routes registered successfully")
}

// HandleStatus handles GET /api/dhcp/status
func (h *DHCPHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling DHCP status request")
	
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	status, err := h.dhcpServer.GetStatus(ctx)
	if err != nil {
		h.logger.Error("Failed to get DHCP status", slog.String("error", err.Error()))
		h.sendError(w, http.StatusInternalServerError, "Failed to get DHCP status")
		return
	}
	
	statistics, err := h.dhcpServer.GetStatistics(ctx)
	if err != nil {
		h.logger.Error("Failed to get DHCP statistics", slog.String("error", err.Error()))
		h.sendError(w, http.StatusInternalServerError, "Failed to get DHCP statistics")
		return
	}
	
	response := DHCPStatusResponse{
		Status:     status,
		Statistics: statistics,
		Timestamp:  time.Now().Format(time.RFC3339),
	}
	
	h.sendJSON(w, response)
}

// HandleLeases handles GET /api/dhcp/leases
func (h *DHCPHandler) HandleLeases(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling DHCP leases request")
	
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	leases, err := h.dhcpServer.GetLeases(ctx)
	if err != nil {
		h.logger.Error("Failed to get DHCP leases", slog.String("error", err.Error()))
		h.sendError(w, http.StatusInternalServerError, "Failed to get DHCP leases")
		return
	}
	
	// Count active and expired leases
	activeCount := 0
	expiredCount := 0
	for _, lease := range leases {
		switch lease.State {
		case types.LeaseStateActive:
			activeCount++
		case types.LeaseStateExpired:
			expiredCount++
		}
	}
	
	response := DHCPLeasesResponse{
		Leases:    leases,
		Total:     len(leases),
		Active:    activeCount,
		Expired:   expiredCount,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	h.sendJSON(w, response)
}

// HandleReservations handles GET /api/dhcp/reservations
func (h *DHCPHandler) HandleReservations(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling DHCP reservations request")
	
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	// Get lease manager to access reservations
	// For now, we'll return an empty response since we need access to the lease manager
	// In a full implementation, the DHCP server would expose reservation methods
	reservations := []types.DHCPReservation{}
	enabledCount := 0
	
	response := DHCPReservationsResponse{
		Reservations: reservations,
		Total:        len(reservations),
		Enabled:      enabledCount,
		Timestamp:    time.Now().Format(time.RFC3339),
	}
	
	h.sendJSON(w, response)
}

// HandleLeaseAction handles actions on individual leases
func (h *DHCPHandler) HandleLeaseAction(w http.ResponseWriter, r *http.Request) {
	// Extract lease ID from URL path
	path := r.URL.Path[len("/api/dhcp/lease/"):]
	if path == "" {
		h.sendError(w, http.StatusBadRequest, "Missing lease identifier")
		return
	}
	
	h.logger.Debug("Handling DHCP lease action", 
		slog.String("method", r.Method),
		slog.String("lease_id", path))
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	switch r.Method {
	case http.MethodGet:
		// Get specific lease
		lease, err := h.dhcpServer.GetLease(ctx, path)
		if err != nil {
			h.logger.Error("Failed to get DHCP lease", 
				slog.String("lease_id", path),
				slog.String("error", err.Error()))
			h.sendError(w, http.StatusNotFound, "Lease not found")
			return
		}
		h.sendJSON(w, lease)
		
	case http.MethodDelete:
		// Release lease (simplified - would need proper implementation)
		h.sendError(w, http.StatusNotImplemented, "Lease release not yet implemented")
		
	default:
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleReservationAction handles actions on DHCP reservations
func (h *DHCPHandler) HandleReservationAction(w http.ResponseWriter, r *http.Request) {
	// Extract MAC address from URL path
	path := r.URL.Path[len("/api/dhcp/reservation/"):]
	if path == "" {
		h.sendError(w, http.StatusBadRequest, "Missing MAC address")
		return
	}
	
	h.logger.Debug("Handling DHCP reservation action",
		slog.String("method", r.Method),
		slog.String("mac", path))
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	switch r.Method {
	case http.MethodPost:
		// Create new reservation
		var reservation types.DHCPReservation
		if err := json.NewDecoder(r.Body).Decode(&reservation); err != nil {
			h.sendError(w, http.StatusBadRequest, "Invalid reservation data")
			return
		}
		
		if err := h.dhcpServer.CreateReservation(ctx, &reservation); err != nil {
			h.logger.Error("Failed to create DHCP reservation",
				slog.String("mac", reservation.MAC),
				slog.String("error", err.Error()))
			h.sendError(w, http.StatusInternalServerError, "Failed to create reservation")
			return
		}
		
		h.sendJSON(w, map[string]string{"status": "created"})
		
	case http.MethodDelete:
		// Delete reservation
		if err := h.dhcpServer.DeleteReservation(ctx, path); err != nil {
			h.logger.Error("Failed to delete DHCP reservation",
				slog.String("mac", path),
				slog.String("error", err.Error()))
			h.sendError(w, http.StatusInternalServerError, "Failed to delete reservation")
			return
		}
		
		h.sendJSON(w, map[string]string{"status": "deleted"})
		
	default:
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleDHCPPage handles the DHCP web interface page
func (h *DHCPHandler) HandleDHCPPage(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling DHCP page request")
	
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	// For now, return a simple HTML page
	// In a full implementation, this would render a proper template
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>DHCP Server Management</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .section { margin: 20px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        .status { background-color: #f0f8f0; }
        .leases { background-color: #f8f8f0; }
        .reservations { background-color: #f0f0f8; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { padding: 8px 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f5f5f5; }
        .btn { padding: 8px 16px; margin: 4px; border: none; border-radius: 3px; cursor: pointer; }
        .btn-primary { background-color: #007bff; color: white; }
        .btn-danger { background-color: #dc3545; color: white; }
        .loading { text-align: center; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌐 DHCP Server Management</h1>
        
        <div class="section status">
            <h2>Server Status</h2>
            <div id="status-content" class="loading">Loading...</div>
        </div>
        
        <div class="section leases">
            <h2>Active Leases</h2>
            <div id="leases-content" class="loading">Loading...</div>
        </div>
        
        <div class="section reservations">
            <h2>IP Reservations</h2>
            <div id="reservations-content" class="loading">Loading...</div>
        </div>
    </div>
    
    <script>
        // Fetch and display DHCP status
        fetch('/api/dhcp/status')
            .then(response => response.json())
            .then(data => {
                const content = document.getElementById('status-content');
                content.innerHTML = '' +
                    '<p><strong>Running:</strong> ' + (data.status.running ? 'Yes' : 'No') + '</p>' +
                    '<p><strong>Interface:</strong> ' + data.status.interface + '</p>' +
                    '<p><strong>Listen Address:</strong> ' + data.status.listen_address + '</p>' +
                    '<p><strong>Pool:</strong> ' + data.status.pool_info.start_ip + ' - ' + data.status.pool_info.end_ip + '</p>' +
                    '<p><strong>Pool Utilization:</strong> ' + data.status.pool_info.utilization_rate.toFixed(1) + '%</p>' +
                    '<p><strong>Total Requests:</strong> ' + data.statistics.total_requests + '</p>';
            })
            .catch(error => {
                document.getElementById('status-content').innerHTML = '<p>Error loading status: ' + error.message + '</p>';
            });
        
        // Fetch and display DHCP leases
        fetch('/api/dhcp/leases')
            .then(response => response.json())
            .then(data => {
                const content = document.getElementById('leases-content');
                if (data.leases.length === 0) {
                    content.innerHTML = '<p>No active leases</p>';
                    return;
                }
                
                let html = '<table><tr><th>IP Address</th><th>MAC Address</th><th>Hostname</th><th>State</th><th>Start Time</th><th>End Time</th><th>Actions</th></tr>';
                data.leases.forEach(lease => {
                    html += '<tr>' +
                        '<td>' + lease.ip + '</td>' +
                        '<td>' + lease.mac + '</td>' +
                        '<td>' + (lease.hostname || '-') + '</td>' +
                        '<td>' + lease.state + '</td>' +
                        '<td>' + new Date(lease.start_time).toLocaleString() + '</td>' +
                        '<td>' + new Date(lease.end_time).toLocaleString() + '</td>' +
                        '<td><button class="btn btn-danger" onclick="releaseLease(\'' + lease.id + '\')">Release</button></td>' +
                        '</tr>';
                });
                html += '</table>';
                content.innerHTML = html;
            })
            .catch(error => {
                document.getElementById('leases-content').innerHTML = '<p>Error loading leases: ' + error.message + '</p>';
            });
        
        // Fetch and display DHCP reservations
        fetch('/api/dhcp/reservations')
            .then(response => response.json())
            .then(data => {
                const content = document.getElementById('reservations-content');
                if (data.reservations.length === 0) {
                    content.innerHTML = '<p>No IP reservations configured</p>';
                    return;
                }
                
                let html = '<table><tr><th>MAC Address</th><th>IP Address</th><th>Hostname</th><th>Description</th><th>Enabled</th><th>Actions</th></tr>';
                data.reservations.forEach(reservation => {
                    html += '<tr>' +
                        '<td>' + reservation.mac + '</td>' +
                        '<td>' + reservation.ip + '</td>' +
                        '<td>' + (reservation.hostname || '-') + '</td>' +
                        '<td>' + (reservation.description || '-') + '</td>' +
                        '<td>' + (reservation.enabled ? 'Yes' : 'No') + '</td>' +
                        '<td><button class="btn btn-danger" onclick="deleteReservation(\'' + reservation.mac + '\')">Delete</button></td>' +
                        '</tr>';
                });
                html += '</table>';
                content.innerHTML = html;
            })
            .catch(error => {
                document.getElementById('reservations-content').innerHTML = '<p>Error loading reservations: ' + error.message + '</p>';
            });
        
        // Release a lease
        function releaseLease(leaseId) {
            if (confirm('Are you sure you want to release this lease?')) {
                fetch('/api/dhcp/lease/' + leaseId, { method: 'DELETE' })
                    .then(response => {
                        if (response.ok) {
                            location.reload();
                        } else {
                            alert('Failed to release lease');
                        }
                    })
                    .catch(error => alert('Error: ' + error.message));
            }
        }
        
        // Delete a reservation
        function deleteReservation(mac) {
            if (confirm('Are you sure you want to delete this reservation?')) {
                fetch('/api/dhcp/reservation/' + mac, { method: 'DELETE' })
                    .then(response => {
                        if (response.ok) {
                            location.reload();
                        } else {
                            alert('Failed to delete reservation');
                        }
                    })
                    .catch(error => alert('Error: ' + error.message));
            }
        }
        
        // Auto-refresh every 30 seconds
        setInterval(() => {
            location.reload();
        }, 30000);
    </script>
</body>
</html>
`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// Helper methods

func (h *DHCPHandler) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *DHCPHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}