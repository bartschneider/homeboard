package db

import (
	"database/sql"
	"fmt"
)

// ClientRepository handles client database operations
type ClientRepository struct {
	db *Database
}

// NewClientRepository creates a new client repository
func NewClientRepository(db *Database) *ClientRepository {
	return &ClientRepository{db: db}
}

// GetAll returns all clients
func (r *ClientRepository) GetAll() ([]Client, error) {
	query := `
		SELECT c.id, c.ip_address, c.name, c.user_agent, c.last_seen, 
		       c.assigned_dashboard_id, c.created_at, c.updated_at,
		       d.name as dashboard_name
		FROM clients c
		LEFT JOIN dashboards d ON c.assigned_dashboard_id = d.id
		ORDER BY c.last_seen DESC
	`

	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var client Client
		var dashboardName sql.NullString

		err := rows.Scan(
			&client.ID, &client.IPAddress, &client.Name, &client.UserAgent,
			&client.LastSeen, &client.AssignedDashboardID, &client.CreatedAt,
			&client.UpdatedAt, &dashboardName,
		)
		if err != nil {
			return nil, err
		}

		clients = append(clients, client)
	}

	return clients, rows.Err()
}

// GetByID returns a client by ID
func (r *ClientRepository) GetByID(id int) (*Client, error) {
	query := `
		SELECT id, ip_address, name, user_agent, last_seen, 
		       assigned_dashboard_id, created_at, updated_at
		FROM clients WHERE id = ?
	`

	var client Client
	err := r.db.conn.QueryRow(query, id).Scan(
		&client.ID, &client.IPAddress, &client.Name, &client.UserAgent,
		&client.LastSeen, &client.AssignedDashboardID, &client.CreatedAt,
		&client.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &client, nil
}

// GetByIP returns a client by IP address
func (r *ClientRepository) GetByIP(ipAddress string) (*Client, error) {
	query := `
		SELECT id, ip_address, name, user_agent, last_seen, 
		       assigned_dashboard_id, created_at, updated_at
		FROM clients WHERE ip_address = ?
	`

	var client Client
	err := r.db.conn.QueryRow(query, ipAddress).Scan(
		&client.ID, &client.IPAddress, &client.Name, &client.UserAgent,
		&client.LastSeen, &client.AssignedDashboardID, &client.CreatedAt,
		&client.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &client, nil
}

// AssignDashboard assigns a dashboard to a client
func (r *ClientRepository) AssignDashboard(clientID, dashboardID int) error {
	_, err := r.db.conn.Exec(
		"UPDATE clients SET assigned_dashboard_id = ? WHERE id = ?",
		dashboardID, clientID,
	)
	return err
}

// UpdateLastSeen updates the last seen timestamp for a client
func (r *ClientRepository) UpdateLastSeen(ipAddress, userAgent string) error {
	return r.db.UpdateLastSeen(ipAddress, userAgent)
}

// WidgetRepository handles widget database operations
type WidgetRepository struct {
	db *Database
}

// NewWidgetRepository creates a new widget repository
func NewWidgetRepository(db *Database) *WidgetRepository {
	return &WidgetRepository{db: db}
}

// GetAll returns all widgets
func (r *WidgetRepository) GetAll() ([]Widget, error) {
	query := `
		SELECT id, name, template_type, data_source, api_url, api_headers, data_mapping,
		       rss_config, description, timeout, enabled, created_at, updated_at
		FROM widgets ORDER BY name
	`

	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var widgets []Widget
	for rows.Next() {
		widget, err := r.scanWidget(rows)
		if err != nil {
			return nil, err
		}
		widgets = append(widgets, widget)
	}

	return widgets, rows.Err()
}

// GetByID returns a widget by ID
func (r *WidgetRepository) GetByID(id int) (*Widget, error) {
	query := `
		SELECT id, name, template_type, data_source, api_url, api_headers, data_mapping,
		       rss_config, description, timeout, enabled, created_at, updated_at
		FROM widgets WHERE id = ?
	`

	row := r.db.conn.QueryRow(query, id)
	widget, err := r.scanWidget(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &widget, nil
}

// Create creates a new widget
func (r *WidgetRepository) Create(widget *Widget) error {
	if err := widget.Validate(); err != nil {
		return err
	}

	// Set default data source if not specified
	if widget.DataSource == "" {
		widget.DataSource = "api"
	}

	apiHeadersJSON, err := widget.APIHeadersJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize API headers: %w", err)
	}

	dataMappingJSON, err := widget.DataMappingJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize data mapping: %w", err)
	}

	rssConfigJSON, err := widget.RSSConfigJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize RSS config: %w", err)
	}

	query := `
		INSERT INTO widgets (name, template_type, data_source, api_url, api_headers, data_mapping,
		                    rss_config, description, timeout, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.conn.Exec(query,
		widget.Name, widget.TemplateType, widget.DataSource, widget.APIURL, apiHeadersJSON,
		dataMappingJSON, rssConfigJSON, widget.Description, widget.Timeout, widget.Enabled,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	widget.ID = int(id)
	return nil
}

// Update updates an existing widget
func (r *WidgetRepository) Update(widget *Widget) error {
	if err := widget.Validate(); err != nil {
		return err
	}

	apiHeadersJSON, err := widget.APIHeadersJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize API headers: %w", err)
	}

	dataMappingJSON, err := widget.DataMappingJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize data mapping: %w", err)
	}

	query := `
		UPDATE widgets 
		SET name = ?, template_type = ?, api_url = ?, api_headers = ?, 
		    data_mapping = ?, description = ?, timeout = ?, enabled = ?
		WHERE id = ?
	`

	_, err = r.db.conn.Exec(query,
		widget.Name, widget.TemplateType, widget.APIURL, apiHeadersJSON,
		dataMappingJSON, widget.Description, widget.Timeout, widget.Enabled,
		widget.ID,
	)

	return err
}

// Delete deletes a widget
func (r *WidgetRepository) Delete(id int) error {
	_, err := r.db.conn.Exec("DELETE FROM widgets WHERE id = ?", id)
	return err
}

// scanWidget scans a database row into a Widget struct
func (r *WidgetRepository) scanWidget(scanner interface{}) (Widget, error) {
	type Scanner interface {
		Scan(dest ...interface{}) error
	}

	var widget Widget
	var apiHeadersJSON, dataMappingJSON, rssConfigJSON string

	err := scanner.(Scanner).Scan(
		&widget.ID, &widget.Name, &widget.TemplateType, &widget.DataSource, &widget.APIURL,
		&apiHeadersJSON, &dataMappingJSON, &rssConfigJSON, &widget.Description,
		&widget.Timeout, &widget.Enabled, &widget.CreatedAt, &widget.UpdatedAt,
	)
	if err != nil {
		return widget, err
	}

	// Deserialize JSON fields
	if err := widget.SetAPIHeadersFromJSON(apiHeadersJSON); err != nil {
		return widget, fmt.Errorf("failed to deserialize API headers: %w", err)
	}
	if err := widget.SetDataMappingFromJSON(dataMappingJSON); err != nil {
		return widget, fmt.Errorf("failed to deserialize data mapping: %w", err)
	}
	if err := widget.SetRSSConfigFromJSON(rssConfigJSON); err != nil {
		return widget, fmt.Errorf("failed to deserialize RSS config: %w", err)
	}

	return widget, nil
}

// DashboardRepository handles dashboard database operations
type DashboardRepository struct {
	db *Database
}

// NewDashboardRepository creates a new dashboard repository
func NewDashboardRepository(db *Database) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// GetAll returns all dashboards
func (r *DashboardRepository) GetAll() ([]Dashboard, error) {
	query := `
		SELECT id, name, description, is_default, created_at, updated_at
		FROM dashboards ORDER BY name
	`

	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dashboards []Dashboard
	for rows.Next() {
		var dashboard Dashboard
		err := rows.Scan(
			&dashboard.ID, &dashboard.Name, &dashboard.Description,
			&dashboard.IsDefault, &dashboard.CreatedAt, &dashboard.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		dashboards = append(dashboards, dashboard)
	}

	return dashboards, rows.Err()
}

// GetByID returns a dashboard by ID with its widgets
func (r *DashboardRepository) GetByID(id int) (*Dashboard, error) {
	query := `
		SELECT id, name, description, is_default, created_at, updated_at
		FROM dashboards WHERE id = ?
	`

	var dashboard Dashboard
	err := r.db.conn.QueryRow(query, id).Scan(
		&dashboard.ID, &dashboard.Name, &dashboard.Description,
		&dashboard.IsDefault, &dashboard.CreatedAt, &dashboard.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load dashboard widgets
	widgets, err := r.GetDashboardWidgets(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load dashboard widgets: %w", err)
	}
	dashboard.Widgets = widgets

	return &dashboard, nil
}

// GetDashboardWidgets returns widgets for a specific dashboard
func (r *DashboardRepository) GetDashboardWidgets(dashboardID int) ([]DashboardWidget, error) {
	query := `
		SELECT dw.id, dw.dashboard_id, dw.widget_id, dw.display_order,
		       dw.grid_x, dw.grid_y, dw.grid_width, dw.grid_height,
		       w.name, w.template_type, w.api_url, w.api_headers, w.data_mapping,
		       w.description, w.timeout, w.enabled, w.created_at, w.updated_at
		FROM dashboard_widgets dw
		JOIN widgets w ON dw.widget_id = w.id
		WHERE dw.dashboard_id = ?
		ORDER BY dw.display_order
	`

	rows, err := r.db.conn.Query(query, dashboardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dashboardWidgets []DashboardWidget
	for rows.Next() {
		var dw DashboardWidget
		var widget Widget
		var apiHeadersJSON, dataMappingJSON string

		err := rows.Scan(
			&dw.ID, &dw.DashboardID, &dw.WidgetID, &dw.DisplayOrder,
			&dw.GridX, &dw.GridY, &dw.GridWidth, &dw.GridHeight,
			&widget.Name, &widget.TemplateType, &widget.APIURL,
			&apiHeadersJSON, &dataMappingJSON, &widget.Description,
			&widget.Timeout, &widget.Enabled, &widget.CreatedAt, &widget.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		widget.ID = dw.WidgetID

		// Deserialize JSON fields
		if err := widget.SetAPIHeadersFromJSON(apiHeadersJSON); err != nil {
			return nil, fmt.Errorf("failed to deserialize API headers: %w", err)
		}
		if err := widget.SetDataMappingFromJSON(dataMappingJSON); err != nil {
			return nil, fmt.Errorf("failed to deserialize data mapping: %w", err)
		}

		dw.Widget = &widget
		dashboardWidgets = append(dashboardWidgets, dw)
	}

	return dashboardWidgets, rows.Err()
}

// Create creates a new dashboard
func (r *DashboardRepository) Create(dashboard *Dashboard) error {
	if err := dashboard.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO dashboards (name, description, is_default)
		VALUES (?, ?, ?)
	`

	result, err := r.db.conn.Exec(query,
		dashboard.Name, dashboard.Description, dashboard.IsDefault,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	dashboard.ID = int(id)
	return nil
}

// Update updates an existing dashboard
func (r *DashboardRepository) Update(dashboard *Dashboard) error {
	if err := dashboard.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE dashboards 
		SET name = ?, description = ?, is_default = ?
		WHERE id = ?
	`

	_, err := r.db.conn.Exec(query,
		dashboard.Name, dashboard.Description, dashboard.IsDefault, dashboard.ID,
	)

	return err
}

// Delete deletes a dashboard
func (r *DashboardRepository) Delete(id int) error {
	return r.db.WithTransaction(func(tx *sql.Tx) error {
		// Delete dashboard widgets first
		_, err := tx.Exec("DELETE FROM dashboard_widgets WHERE dashboard_id = ?", id)
		if err != nil {
			return err
		}

		// Delete dashboard
		_, err = tx.Exec("DELETE FROM dashboards WHERE id = ?", id)
		return err
	})
}

// AddWidget adds a widget to a dashboard
func (r *DashboardRepository) AddWidget(dashboardID, widgetID, displayOrder int) error {
	query := `
		INSERT INTO dashboard_widgets (dashboard_id, widget_id, display_order, grid_width, grid_height)
		VALUES (?, ?, ?, 1, 1)
	`

	_, err := r.db.conn.Exec(query, dashboardID, widgetID, displayOrder)
	return err
}

// RemoveWidget removes a widget from a dashboard
func (r *DashboardRepository) RemoveWidget(dashboardID, widgetID int) error {
	_, err := r.db.conn.Exec(
		"DELETE FROM dashboard_widgets WHERE dashboard_id = ? AND widget_id = ?",
		dashboardID, widgetID,
	)
	return err
}

// UpdateWidgetOrder updates the display order of widgets in a dashboard
func (r *DashboardRepository) UpdateWidgetOrder(dashboardID int, widgetOrders []struct {
	WidgetID     int `json:"widget_id"`
	DisplayOrder int `json:"display_order"`
}) error {
	return r.db.WithTransaction(func(tx *sql.Tx) error {
		for _, order := range widgetOrders {
			_, err := tx.Exec(
				"UPDATE dashboard_widgets SET display_order = ? WHERE dashboard_id = ? AND widget_id = ?",
				order.DisplayOrder, dashboardID, order.WidgetID,
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// GetDefault returns the default dashboard
func (r *DashboardRepository) GetDefault() (*Dashboard, error) {
	query := `
		SELECT id, name, description, is_default, created_at, updated_at
		FROM dashboards WHERE is_default = true LIMIT 1
	`

	var dashboard Dashboard
	err := r.db.conn.QueryRow(query).Scan(
		&dashboard.ID, &dashboard.Name, &dashboard.Description,
		&dashboard.IsDefault, &dashboard.CreatedAt, &dashboard.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load dashboard widgets
	widgets, err := r.GetDashboardWidgets(dashboard.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load dashboard widgets: %w", err)
	}
	dashboard.Widgets = widgets

	return &dashboard, nil
}
