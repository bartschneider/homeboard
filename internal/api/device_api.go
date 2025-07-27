package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/bartosz/homeboard/internal/db"
	"github.com/gorilla/mux"
)

// Device represents a registered device
type Device struct {
	ID           string    `json:"device_id" db:"device_id"`
	Name         string    `json:"device_name" db:"device_name"`
	Type         string    `json:"device_type" db:"device_type"`
	Capabilities []string  `json:"capabilities" db:"capabilities"`
	DashboardURL string    `json:"dashboard_url,omitempty" db:"dashboard_url"`
	LastSeen     time.Time `json:"last_seen" db:"last_seen"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// DeviceRegistrationRequest represents a device registration request
type DeviceRegistrationRequest struct {
	DeviceID     string   `json:"device_id"`
	DeviceName   string   `json:"device_name"`
	DeviceType   string   `json:"device_type"`
	Capabilities []string `json:"capabilities"`
}

// DeviceAssignmentResponse represents a dashboard assignment response
type DeviceAssignmentResponse struct {
	DeviceID        string `json:"device_id"`
	DashboardURL    string `json:"dashboard_url"`
	RefreshInterval int    `json:"refresh_interval"`
}

// DeviceHandler handles device-related API endpoints
type DeviceHandler struct {
	database *db.Database
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(database *db.Database) *DeviceHandler {
	return &DeviceHandler{
		database: database,
	}
}

// RegisterDevice handles device registration
func (h *DeviceHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	var req DeviceRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.DeviceID == "" || req.DeviceName == "" || req.DeviceType == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Check if device already exists
	existingDevice, err := h.getDeviceByID(req.DeviceID)
	if err == nil && existingDevice != nil {
		// Update existing device
		existingDevice.Name = req.DeviceName
		existingDevice.Type = req.DeviceType
		existingDevice.Capabilities = req.Capabilities
		existingDevice.LastSeen = time.Now()
		existingDevice.IsActive = true
		existingDevice.UpdatedAt = time.Now()

		if err := h.updateDevice(existingDevice); err != nil {
			log.Printf("Error updating device: %v", err)
			http.Error(w, "Failed to update device", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(existingDevice)
		return
	}

	// Create new device
	device := &Device{
		ID:           req.DeviceID,
		Name:         req.DeviceName,
		Type:         req.DeviceType,
		Capabilities: req.Capabilities,
		DashboardURL: "", // Will be assigned later
		LastSeen:     time.Now(),
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.createDevice(device); err != nil {
		log.Printf("Error creating device: %v", err)
		http.Error(w, "Failed to register device", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(device)
}

// GetDeviceDashboard returns the assigned dashboard for a device
func (h *DeviceHandler) GetDeviceDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["device_id"]

	if deviceID == "" {
		http.Error(w, "Device ID is required", http.StatusBadRequest)
		return
	}

	device, err := h.getDeviceByID(deviceID)
	if err != nil {
		log.Printf("Error getting device: %v", err)
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// Update last seen timestamp
	device.LastSeen = time.Now()
	h.updateDevice(device)

	// Determine dashboard URL
	dashboardURL := device.DashboardURL
	if dashboardURL == "" {
		// Default to main dashboard if no specific assignment
		dashboardURL = "/"
	}

	response := DeviceAssignmentResponse{
		DeviceID:        device.ID,
		DashboardURL:    dashboardURL,
		RefreshInterval: 900, // 15 minutes default
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListDevices returns all registered devices
func (h *DeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.getAllDevices()
	if err != nil {
		log.Printf("Error getting devices: %v", err)
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

// GetDevice returns a specific device
func (h *DeviceHandler) GetDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["device_id"]

	if deviceID == "" {
		http.Error(w, "Device ID is required", http.StatusBadRequest)
		return
	}

	device, err := h.getDeviceByID(deviceID)
	if err != nil {
		log.Printf("Error getting device: %v", err)
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(device)
}

// AssignDashboard assigns a specific dashboard to a device
func (h *DeviceHandler) AssignDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["device_id"]

	if deviceID == "" {
		http.Error(w, "Device ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		DashboardURL string `json:"dashboard_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	device, err := h.getDeviceByID(deviceID)
	if err != nil {
		log.Printf("Error getting device: %v", err)
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// Update dashboard assignment
	device.DashboardURL = req.DashboardURL
	device.UpdatedAt = time.Now()

	if err := h.updateDevice(device); err != nil {
		log.Printf("Error updating device: %v", err)
		http.Error(w, "Failed to assign dashboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(device)
}

// Database helper methods
func (h *DeviceHandler) createDevice(device *Device) error {
	capabilitiesJSON, _ := json.Marshal(device.Capabilities)

	query := `
		INSERT INTO devices (device_id, device_name, device_type, capabilities, dashboard_url, last_seen, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := h.database.DB.Exec(query,
		device.ID, device.Name, device.Type, string(capabilitiesJSON),
		device.DashboardURL, device.LastSeen, device.IsActive,
		device.CreatedAt, device.UpdatedAt)

	return err
}

func (h *DeviceHandler) getDeviceByID(deviceID string) (*Device, error) {
	query := `
		SELECT device_id, device_name, device_type, capabilities, dashboard_url, last_seen, is_active, created_at, updated_at
		FROM devices WHERE device_id = ?
	`

	var device Device
	var capabilitiesJSON string

	err := h.database.DB.QueryRow(query, deviceID).Scan(
		&device.ID, &device.Name, &device.Type, &capabilitiesJSON,
		&device.DashboardURL, &device.LastSeen, &device.IsActive,
		&device.CreatedAt, &device.UpdatedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(capabilitiesJSON), &device.Capabilities)
	return &device, nil
}

func (h *DeviceHandler) getAllDevices() ([]Device, error) {
	query := `
		SELECT device_id, device_name, device_type, capabilities, dashboard_url, last_seen, is_active, created_at, updated_at
		FROM devices ORDER BY created_at DESC
	`

	rows, err := h.database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var device Device
		var capabilitiesJSON string

		err := rows.Scan(
			&device.ID, &device.Name, &device.Type, &capabilitiesJSON,
			&device.DashboardURL, &device.LastSeen, &device.IsActive,
			&device.CreatedAt, &device.UpdatedAt)

		if err != nil {
			continue
		}

		json.Unmarshal([]byte(capabilitiesJSON), &device.Capabilities)
		devices = append(devices, device)
	}

	return devices, nil
}

func (h *DeviceHandler) updateDevice(device *Device) error {
	capabilitiesJSON, _ := json.Marshal(device.Capabilities)

	query := `
		UPDATE devices SET
			device_name = ?, device_type = ?, capabilities = ?, dashboard_url = ?,
			last_seen = ?, is_active = ?, updated_at = ?
		WHERE device_id = ?
	`

	_, err := h.database.DB.Exec(query,
		device.Name, device.Type, string(capabilitiesJSON), device.DashboardURL,
		device.LastSeen, device.IsActive, device.UpdatedAt, device.ID)

	return err
}

// Health check endpoint
func (h *DeviceHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"service":   "homeboard-api",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
