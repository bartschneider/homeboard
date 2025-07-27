package dto

import (
	"time"
)

// CreateDashboardRequest represents a request to create a new dashboard
type CreateDashboardRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=500"`
	IsDefault   bool   `json:"is_default"`
}

// UpdateDashboardRequest represents a request to update an existing dashboard
type UpdateDashboardRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	IsDefault   *bool   `json:"is_default,omitempty"`
}

// DashboardResponse represents the response when returning dashboard data
type DashboardResponse struct {
	ID          int                       `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	IsDefault   bool                      `json:"is_default"`
	Widgets     []DashboardWidgetResponse `json:"widgets"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

// DashboardListResponse represents a paginated list of dashboards
type DashboardListResponse struct {
	Dashboards []DashboardSummaryResponse `json:"dashboards"`
	Pagination PaginationResponse         `json:"pagination"`
}

// DashboardSummaryResponse represents a simplified dashboard for list views
type DashboardSummaryResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsDefault   bool      `json:"is_default"`
	WidgetCount int       `json:"widget_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// DashboardWidgetResponse represents a widget within a dashboard
type DashboardWidgetResponse struct {
	ID           int                    `json:"id"`
	DashboardID  int                    `json:"dashboard_id"`
	WidgetID     int                    `json:"widget_id"`
	DisplayOrder int                    `json:"display_order"`
	GridPosition GridPositionDTO        `json:"grid_position"`
	Widget       *WidgetSummaryResponse `json:"widget,omitempty"`
}

// GridPositionDTO represents widget position and size in a grid layout
type GridPositionDTO struct {
	X      int `json:"x" validate:"min=0"`
	Y      int `json:"y" validate:"min=0"`
	Width  int `json:"width" validate:"min=1,max=12"`
	Height int `json:"height" validate:"min=1,max=12"`
}

// AddWidgetToDashboardRequest represents a request to add a widget to a dashboard
type AddWidgetToDashboardRequest struct {
	WidgetID     int             `json:"widget_id" validate:"required,min=1"`
	DisplayOrder int             `json:"display_order" validate:"min=0"`
	GridPosition GridPositionDTO `json:"grid_position"`
}

// ReorderDashboardWidgetsRequest represents a request to reorder widgets in a dashboard
type ReorderDashboardWidgetsRequest struct {
	Widgets []DashboardWidgetOrderDTO `json:"widgets" validate:"required,dive"`
}

// DashboardWidgetOrderDTO represents the new order for dashboard widgets
type DashboardWidgetOrderDTO struct {
	ID           int             `json:"id" validate:"required,min=1"`
	DisplayOrder int             `json:"display_order" validate:"min=0"`
	GridPosition GridPositionDTO `json:"grid_position"`
}

// DashboardStatsResponse represents statistics about dashboards
type DashboardStatsResponse struct {
	TotalDashboards  int                        `json:"total_dashboards"`
	TotalWidgets     int                        `json:"total_widgets"`
	DefaultDashboard *DashboardSummaryResponse  `json:"default_dashboard,omitempty"`
	RecentDashboards []DashboardSummaryResponse `json:"recent_dashboards"`
}
