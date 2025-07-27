package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database connection and operations
type Database struct {
	conn   *sql.DB
	dbPath string
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	// Ensure the database directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open SQLite database
	conn, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &Database{
		conn:   conn,
		dbPath: dbPath,
	}

	// Initialize the database schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Printf("Database initialized successfully at: %s", dbPath)
	return db, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// initSchema creates the database tables if they don't exist
func (db *Database) initSchema() error {
	schema := `
	-- Clients table
	CREATE TABLE IF NOT EXISTS clients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip_address TEXT UNIQUE NOT NULL,
		name TEXT DEFAULT '',
		user_agent TEXT DEFAULT '',
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		assigned_dashboard_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (assigned_dashboard_id) REFERENCES dashboards(id) ON DELETE SET NULL
	);

	-- Widgets table
	CREATE TABLE IF NOT EXISTS widgets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		template_type TEXT NOT NULL,
		data_source TEXT DEFAULT 'api',
		api_url TEXT DEFAULT '',
		api_headers TEXT DEFAULT '{}',
		data_mapping TEXT DEFAULT '{}',
		rss_config TEXT DEFAULT '',
		description TEXT DEFAULT '',
		timeout INTEGER DEFAULT 30,
		enabled BOOLEAN DEFAULT true,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Dashboards table
	CREATE TABLE IF NOT EXISTS dashboards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		is_default BOOLEAN DEFAULT false,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Dashboard widgets join table
	CREATE TABLE IF NOT EXISTS dashboard_widgets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		dashboard_id INTEGER NOT NULL,
		widget_id INTEGER NOT NULL,
		display_order INTEGER DEFAULT 0,
		grid_x INTEGER DEFAULT 0,
		grid_y INTEGER DEFAULT 0,
		grid_width INTEGER DEFAULT 1,
		grid_height INTEGER DEFAULT 1,
		FOREIGN KEY (dashboard_id) REFERENCES dashboards(id) ON DELETE CASCADE,
		FOREIGN KEY (widget_id) REFERENCES widgets(id) ON DELETE CASCADE,
		UNIQUE(dashboard_id, widget_id)
	);

	-- Indexes for better performance
	CREATE INDEX IF NOT EXISTS idx_clients_ip ON clients(ip_address);
	CREATE INDEX IF NOT EXISTS idx_clients_last_seen ON clients(last_seen);
	CREATE INDEX IF NOT EXISTS idx_widgets_enabled ON widgets(enabled);
	CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_dashboard ON dashboard_widgets(dashboard_id);
	CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_order ON dashboard_widgets(dashboard_id, display_order);

	-- Triggers for updated_at timestamps
	CREATE TRIGGER IF NOT EXISTS update_clients_timestamp 
		AFTER UPDATE ON clients
		BEGIN
			UPDATE clients SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

	CREATE TRIGGER IF NOT EXISTS update_widgets_timestamp 
		AFTER UPDATE ON widgets
		BEGIN
			UPDATE widgets SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

	CREATE TRIGGER IF NOT EXISTS update_dashboards_timestamp 
		AFTER UPDATE ON dashboards
		BEGIN
			UPDATE dashboards SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`

	if _, err := db.conn.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Create default dashboard if none exists
	if err := db.createDefaultDashboard(); err != nil {
		return fmt.Errorf("failed to create default dashboard: %w", err)
	}

	return nil
}

// createDefaultDashboard creates a default dashboard if none exists
func (db *Database) createDefaultDashboard() error {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM dashboards").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = db.conn.Exec(`
			INSERT INTO dashboards (name, description, is_default)
			VALUES (?, ?, ?)
		`, "Default Dashboard", "Default dashboard for new clients", true)
		return err
	}

	return nil
}

// Transaction helper
func (db *Database) WithTransaction(fn func(*sql.Tx) error) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// UpdateLastSeen updates the last_seen timestamp for a client
func (db *Database) UpdateLastSeen(ipAddress, userAgent string) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		INSERT INTO clients (ip_address, user_agent, last_seen, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(ip_address) DO UPDATE SET
			last_seen = ?,
			user_agent = ?,
			updated_at = ?
	`, ipAddress, userAgent, now, now, now, now, userAgent, now)

	return err
}

// GetHealth returns database health information
func (db *Database) GetHealth() map[string]interface{} {
	stats := db.conn.Stats()

	var totalClients, totalWidgets, totalDashboards int
	db.conn.QueryRow("SELECT COUNT(*) FROM clients").Scan(&totalClients)
	db.conn.QueryRow("SELECT COUNT(*) FROM widgets").Scan(&totalWidgets)
	db.conn.QueryRow("SELECT COUNT(*) FROM dashboards").Scan(&totalDashboards)

	return map[string]interface{}{
		"status":           "healthy",
		"database_path":    db.dbPath,
		"open_connections": stats.OpenConnections,
		"in_use":           stats.InUse,
		"idle":             stats.Idle,
		"total_clients":    totalClients,
		"total_widgets":    totalWidgets,
		"total_dashboards": totalDashboards,
	}
}

// Backup creates a backup of the database
func (db *Database) Backup(backupPath string) error {
	// Ensure backup directory exists
	dir := filepath.Dir(backupPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Simple file copy for SQLite backup
	sourceFile, err := os.Open(db.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open source database: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()

	// Copy the file
	_, err = destFile.ReadFrom(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy database: %w", err)
	}

	return nil
}

// Vacuum performs database maintenance
func (db *Database) Vacuum() error {
	_, err := db.conn.Exec("VACUUM")
	return err
}

// GetConnection returns the underlying database connection (for advanced operations)
func (db *Database) GetConnection() *sql.DB {
	return db.conn
}
