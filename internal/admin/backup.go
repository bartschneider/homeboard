package admin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bartosz/homeboard/internal/config"
)

// BackupManager handles configuration backup and restore operations
type BackupManager struct {
	configPath string
	backupDir  string
}

// NewBackupManager creates a new backup manager
func NewBackupManager(configPath string) *BackupManager {
	backupDir := filepath.Join(filepath.Dir(configPath), "backups")
	
	// Ensure backup directory exists
	os.MkdirAll(backupDir, 0755)
	
	return &BackupManager{
		configPath: configPath,
		backupDir:  backupDir,
	}
}

// CreateBackup creates a backup of the current configuration
func (bm *BackupManager) CreateBackup() (string, error) {
	// Read current configuration
	data, err := os.ReadFile(bm.configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read configuration: %w", err)
	}

	// Generate backup ID based on timestamp
	backupID := fmt.Sprintf("backup_%s", time.Now().Format("20060102_150405"))
	backupFilename := fmt.Sprintf("%s.json", backupID)
	backupPath := filepath.Join(bm.backupDir, backupFilename)

	// Write backup file
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	// Create metadata file
	if err := bm.createBackupMetadata(backupID, backupPath, data); err != nil {
		// Log error but don't fail the backup
		fmt.Printf("Warning: failed to create backup metadata: %v\n", err)
	}

	return backupID, nil
}

// RestoreBackup restores configuration from a backup
func (bm *BackupManager) RestoreBackup(backupID string) error {
	backupFilename := fmt.Sprintf("%s.json", backupID)
	backupPath := filepath.Join(bm.backupDir, backupFilename)

	// Check if backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupID)
	}

	// Read backup data
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Validate backup data by trying to parse it
	if _, err := config.LoadConfig(backupPath); err != nil {
		return fmt.Errorf("backup file is corrupted or invalid: %w", err)
	}

	// Create a backup of current configuration before restoring
	currentBackupID, err := bm.CreateBackup()
	if err != nil {
		return fmt.Errorf("failed to backup current configuration before restore: %w", err)
	}

	// Restore configuration
	if err := os.WriteFile(bm.configPath, data, 0644); err != nil {
		// Try to restore the backup we just created
		bm.RestoreBackup(currentBackupID)
		return fmt.Errorf("failed to restore configuration: %w", err)
	}

	return nil
}

// ListBackups returns a list of available backups
func (bm *BackupManager) ListBackups() ([]BackupInfo, error) {
	files, err := os.ReadDir(bm.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Extract backup ID from filename
		backupID := strings.TrimSuffix(file.Name(), ".json")
		
		// Get file info
		info, err := file.Info()
		if err != nil {
			continue
		}

		// Load metadata if available
		metadata := bm.loadBackupMetadata(backupID)

		backup := BackupInfo{
			ID:          backupID,
			Filename:    file.Name(),
			CreatedAt:   info.ModTime(),
			Size:        info.Size(),
			Description: metadata.Description,
		}

		backups = append(backups, backup)
	}

	// Sort backups by creation date (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// DeleteBackup deletes a specific backup
func (bm *BackupManager) DeleteBackup(backupID string) error {
	backupFilename := fmt.Sprintf("%s.json", backupID)
	backupPath := filepath.Join(bm.backupDir, backupFilename)

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// Delete backup file
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup file: %w", err)
	}

	// Delete metadata file if it exists
	metadataPath := filepath.Join(bm.backupDir, fmt.Sprintf("%s.meta", backupID))
	if _, err := os.Stat(metadataPath); err == nil {
		os.Remove(metadataPath) // Ignore errors for metadata cleanup
	}

	return nil
}

// CleanupOldBackups removes backups older than the specified duration
func (bm *BackupManager) CleanupOldBackups(maxAge time.Duration) error {
	backups, err := bm.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	cutoffTime := time.Now().Add(-maxAge)
	deletedCount := 0

	for _, backup := range backups {
		if backup.CreatedAt.Before(cutoffTime) {
			if err := bm.DeleteBackup(backup.ID); err != nil {
				fmt.Printf("Warning: failed to delete old backup %s: %v\n", backup.ID, err)
			} else {
				deletedCount++
			}
		}
	}

	return nil
}

// GetBackupInfo returns detailed information about a specific backup
func (bm *BackupManager) GetBackupInfo(backupID string) (*BackupMetadata, error) {
	backupFilename := fmt.Sprintf("%s.json", backupID)
	backupPath := filepath.Join(bm.backupDir, backupFilename)

	// Check if backup exists
	info, err := os.Stat(backupPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("backup not found: %s", backupID)
	}

	// Calculate checksum
	checksum, err := bm.calculateChecksum(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Load metadata
	metadata := bm.loadBackupMetadata(backupID)
	if metadata == nil {
		metadata = &BackupMetadata{
			ID:        backupID,
			Filename:  backupFilename,
			CreatedAt: info.ModTime(),
		}
	}

	metadata.Size = info.Size()
	metadata.Checksum = checksum

	return metadata, nil
}

// ValidateBackup validates the integrity of a backup file
func (bm *BackupManager) ValidateBackup(backupID string) error {
	backupFilename := fmt.Sprintf("%s.json", backupID)
	backupPath := filepath.Join(bm.backupDir, backupFilename)

	// Check if file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupID)
	}

	// Try to parse the configuration
	if _, err := config.LoadConfig(backupPath); err != nil {
		return fmt.Errorf("backup file is corrupted or invalid: %w", err)
	}

	// Validate checksum if metadata exists
	metadata := bm.loadBackupMetadata(backupID)
	if metadata != nil && metadata.Checksum != "" {
		currentChecksum, err := bm.calculateChecksum(backupPath)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum: %w", err)
		}

		if currentChecksum != metadata.Checksum {
			return fmt.Errorf("backup checksum mismatch: expected %s, got %s", metadata.Checksum, currentChecksum)
		}
	}

	return nil
}

// ExportBackup exports a backup with metadata to a specified path
func (bm *BackupManager) ExportBackup(backupID, exportPath string) error {
	backupFilename := fmt.Sprintf("%s.json", backupID)
	backupPath := filepath.Join(bm.backupDir, backupFilename)

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// Copy backup file
	if err := bm.copyFile(backupPath, exportPath); err != nil {
		return fmt.Errorf("failed to export backup: %w", err)
	}

	return nil
}

// ImportBackup imports a backup from an external file
func (bm *BackupManager) ImportBackup(importPath, description string) (string, error) {
	// Validate the import file
	if _, err := config.LoadConfig(importPath); err != nil {
		return "", fmt.Errorf("import file is not a valid configuration: %w", err)
	}

	// Read import data
	data, err := os.ReadFile(importPath)
	if err != nil {
		return "", fmt.Errorf("failed to read import file: %w", err)
	}

	// Generate backup ID
	backupID := fmt.Sprintf("imported_%s", time.Now().Format("20060102_150405"))
	backupFilename := fmt.Sprintf("%s.json", backupID)
	backupPath := filepath.Join(bm.backupDir, backupFilename)

	// Write backup file
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}

	// Create metadata
	if err := bm.createBackupMetadata(backupID, backupPath, data); err != nil {
		fmt.Printf("Warning: failed to create backup metadata: %v\n", err)
	}

	// Update description if provided
	if description != "" {
		metadata := bm.loadBackupMetadata(backupID)
		if metadata != nil {
			metadata.Description = description
			bm.saveBackupMetadata(backupID, metadata)
		}
	}

	return backupID, nil
}

// Helper methods

// createBackupMetadata creates metadata for a backup
func (bm *BackupManager) createBackupMetadata(backupID, backupPath string, data []byte) error {
	checksum := bm.calculateChecksumFromData(data)
	
	metadata := BackupMetadata{
		ID:          backupID,
		Filename:    filepath.Base(backupPath),
		CreatedAt:   time.Now(),
		Size:        int64(len(data)),
		Checksum:    checksum,
		Description: "Automatic backup",
		Metadata: map[string]string{
			"created_by": "admin_panel",
			"version":    "1.0",
		},
	}

	return bm.saveBackupMetadata(backupID, &metadata)
}

// loadBackupMetadata loads metadata for a backup
func (bm *BackupManager) loadBackupMetadata(backupID string) *BackupMetadata {
	metadataPath := filepath.Join(bm.backupDir, fmt.Sprintf("%s.meta", backupID))
	
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil
	}

	return &metadata
}

// saveBackupMetadata saves metadata for a backup
func (bm *BackupManager) saveBackupMetadata(backupID string, metadata *BackupMetadata) error {
	metadataPath := filepath.Join(bm.backupDir, fmt.Sprintf("%s.meta", backupID))
	
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

// calculateChecksum calculates SHA256 checksum of a file
func (bm *BackupManager) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// calculateChecksumFromData calculates SHA256 checksum of data
func (bm *BackupManager) calculateChecksumFromData(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// copyFile copies a file from source to destination
func (bm *BackupManager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}