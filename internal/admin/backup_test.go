package admin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bartosz/homeboard/internal/config"
)

func TestBackupManager(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	// Create test configuration
	testConfig := &config.Config{
		RefreshInterval: 15,
		ServerPort:      8081,
		Title:           "Test Dashboard",
		Theme: config.Theme{
			FontFamily: "serif",
			FontSize:   "16px",
			Background: "#ffffff",
			Foreground: "#000000",
		},
		Widgets: []config.Widget{
			{
				Name:    "Test Widget",
				Script:  "test.py",
				Enabled: true,
				Timeout: 10,
				Parameters: map[string]interface{}{
					"test_param": "value",
				},
			},
		},
	}

	// Save test configuration
	if err := config.SaveConfig(testConfig, configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Create backup manager
	backupManager := NewBackupManager(configPath)

	t.Run("CreateBackup", func(t *testing.T) {
		backupID, err := backupManager.CreateBackup()
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		if backupID == "" {
			t.Error("Expected non-empty backup ID")
		}

		// Verify backup file exists
		backupPath := filepath.Join(backupManager.backupDir, backupID+".json")
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Errorf("Backup file does not exist: %s", backupPath)
		}

		// Verify backup content
		backupData, err := os.ReadFile(backupPath)
		if err != nil {
			t.Fatalf("Failed to read backup file: %v", err)
		}

		var backupConfig config.Config
		if err := json.Unmarshal(backupData, &backupConfig); err != nil {
			t.Fatalf("Failed to unmarshal backup config: %v", err)
		}

		if backupConfig.Title != testConfig.Title {
			t.Errorf("Expected backup title '%s', got '%s'", testConfig.Title, backupConfig.Title)
		}
	})

	t.Run("ListBackups", func(t *testing.T) {
		// Create multiple backups
		backupIDs := make([]string, 3)
		for i := 0; i < 3; i++ {
			backupID, err := backupManager.CreateBackup()
			if err != nil {
				t.Fatalf("Failed to create backup %d: %v", i, err)
			}
			backupIDs[i] = backupID
			time.Sleep(10 * time.Millisecond) // Ensure different timestamps
		}

		backups, err := backupManager.ListBackups()
		if err != nil {
			t.Fatalf("Failed to list backups: %v", err)
		}

		if len(backups) < 3 {
			t.Errorf("Expected at least 3 backups, got %d", len(backups))
		}

		// Verify backups are sorted by creation date (newest first)
		for i := 1; i < len(backups); i++ {
			if backups[i-1].CreatedAt.Before(backups[i].CreatedAt) {
				t.Error("Backups should be sorted by creation date (newest first)")
			}
		}

		// Verify backup info
		for _, backup := range backups {
			if backup.ID == "" {
				t.Error("Expected non-empty backup ID")
			}
			if backup.Size == 0 {
				t.Error("Expected non-zero backup size")
			}
		}
	})

	t.Run("RestoreBackup", func(t *testing.T) {
		// Create backup
		backupID, err := backupManager.CreateBackup()
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		// Modify original configuration
		modifiedConfig := *testConfig
		modifiedConfig.Title = "Modified Dashboard"
		modifiedConfig.RefreshInterval = 30

		if err := config.SaveConfig(&modifiedConfig, configPath); err != nil {
			t.Fatalf("Failed to save modified config: %v", err)
		}

		// Restore from backup
		if err := backupManager.RestoreBackup(backupID); err != nil {
			t.Fatalf("Failed to restore backup: %v", err)
		}

		// Verify restoration
		restoredConfig, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load restored config: %v", err)
		}

		if restoredConfig.Title != testConfig.Title {
			t.Errorf("Expected restored title '%s', got '%s'", testConfig.Title, restoredConfig.Title)
		}

		if restoredConfig.RefreshInterval != testConfig.RefreshInterval {
			t.Errorf("Expected restored refresh interval %d, got %d", testConfig.RefreshInterval, restoredConfig.RefreshInterval)
		}
	})

	t.Run("DeleteBackup", func(t *testing.T) {
		// Create backup
		backupID, err := backupManager.CreateBackup()
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		// Verify backup exists
		backupPath := filepath.Join(backupManager.backupDir, backupID+".json")
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Error("Backup file should exist before deletion")
		}

		// Delete backup
		if err := backupManager.DeleteBackup(backupID); err != nil {
			t.Fatalf("Failed to delete backup: %v", err)
		}

		// Verify backup is deleted
		if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
			t.Error("Backup file should not exist after deletion")
		}
	})

	t.Run("ValidateBackup", func(t *testing.T) {
		// Create valid backup
		backupID, err := backupManager.CreateBackup()
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		// Validate backup
		if err := backupManager.ValidateBackup(backupID); err != nil {
			t.Errorf("Expected valid backup, but validation failed: %v", err)
		}

		// Test with non-existent backup
		if err := backupManager.ValidateBackup("non_existent_backup"); err == nil {
			t.Error("Expected validation to fail for non-existent backup")
		}

		// Test with corrupted backup
		corruptedBackupID := "corrupted_backup"
		corruptedPath := filepath.Join(backupManager.backupDir, corruptedBackupID+".json")
		if err := os.WriteFile(corruptedPath, []byte("invalid json"), 0644); err != nil {
			t.Fatalf("Failed to create corrupted backup: %v", err)
		}

		if err := backupManager.ValidateBackup(corruptedBackupID); err == nil {
			t.Error("Expected validation to fail for corrupted backup")
		}
	})

	t.Run("GetBackupInfo", func(t *testing.T) {
		// Create backup
		backupID, err := backupManager.CreateBackup()
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		// Get backup info
		info, err := backupManager.GetBackupInfo(backupID)
		if err != nil {
			t.Fatalf("Failed to get backup info: %v", err)
		}

		if info.ID != backupID {
			t.Errorf("Expected backup ID '%s', got '%s'", backupID, info.ID)
		}

		if info.Size == 0 {
			t.Error("Expected non-zero backup size")
		}

		if info.Checksum == "" {
			t.Error("Expected non-empty checksum")
		}

		if info.CreatedAt.IsZero() {
			t.Error("Expected non-zero creation time")
		}
	})

	t.Run("CleanupOldBackups", func(t *testing.T) {
		// Create backups with different ages
		backupIDs := make([]string, 3)
		for i := 0; i < 3; i++ {
			backupID, err := backupManager.CreateBackup()
			if err != nil {
				t.Fatalf("Failed to create backup %d: %v", i, err)
			}
			backupIDs[i] = backupID

			// Artificially age some backups by modifying file timestamps
			if i < 2 {
				backupPath := filepath.Join(backupManager.backupDir, backupID+".json")
				oldTime := time.Now().Add(-2 * time.Hour)
				if err := os.Chtimes(backupPath, oldTime, oldTime); err != nil {
					t.Logf("Warning: Failed to modify backup timestamp: %v", err)
				}
			}
		}

		// Clean up backups older than 1 hour
		if err := backupManager.CleanupOldBackups(1 * time.Hour); err != nil {
			t.Fatalf("Failed to cleanup old backups: %v", err)
		}

		// Verify that old backups are cleaned up
		backups, err := backupManager.ListBackups()
		if err != nil {
			t.Fatalf("Failed to list backups after cleanup: %v", err)
		}

		// Should have at least the recent backup
		recentBackupExists := false
		for _, backup := range backups {
			if backup.ID == backupIDs[2] { // The most recent backup
				recentBackupExists = true
				break
			}
		}

		if !recentBackupExists {
			t.Error("Recent backup should not be cleaned up")
		}
	})

	t.Run("ExportBackup", func(t *testing.T) {
		// Create backup
		backupID, err := backupManager.CreateBackup()
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		// Export backup
		exportPath := filepath.Join(tempDir, "exported_backup.json")
		if err := backupManager.ExportBackup(backupID, exportPath); err != nil {
			t.Fatalf("Failed to export backup: %v", err)
		}

		// Verify exported file exists
		if _, err := os.Stat(exportPath); os.IsNotExist(err) {
			t.Error("Exported backup file should exist")
		}

		// Verify exported content is valid
		if _, err := config.LoadConfig(exportPath); err != nil {
			t.Errorf("Exported backup should be valid config: %v", err)
		}
	})

	t.Run("ImportBackup", func(t *testing.T) {
		// Create a configuration file to import
		importConfig := &config.Config{
			RefreshInterval: 20,
			ServerPort:      8082,
			Title:           "Imported Dashboard",
			Theme: config.Theme{
				FontFamily: "sans-serif",
				FontSize:   "18px",
				Background: "#f0f0f0",
				Foreground: "#111111",
			},
			Widgets: []config.Widget{
				{
					Name:    "Imported Widget",
					Script:  "imported.py",
					Enabled: true,
					Timeout: 15,
				},
			},
		}

		importPath := filepath.Join(tempDir, "import_config.json")
		if err := config.SaveConfig(importConfig, importPath); err != nil {
			t.Fatalf("Failed to save import config: %v", err)
		}

		// Import backup
		description := "Imported test configuration"
		backupID, err := backupManager.ImportBackup(importPath, description)
		if err != nil {
			t.Fatalf("Failed to import backup: %v", err)
		}

		if backupID == "" {
			t.Error("Expected non-empty backup ID")
		}

		// Verify imported backup
		backups, err := backupManager.ListBackups()
		if err != nil {
			t.Fatalf("Failed to list backups: %v", err)
		}

		importedBackupFound := false
		for _, backup := range backups {
			if backup.ID == backupID {
				importedBackupFound = true
				if backup.Description != description {
					t.Errorf("Expected description '%s', got '%s'", description, backup.Description)
				}
				break
			}
		}

		if !importedBackupFound {
			t.Error("Imported backup should be found in backup list")
		}

		// Verify backup content
		if err := backupManager.ValidateBackup(backupID); err != nil {
			t.Errorf("Imported backup should be valid: %v", err)
		}
	})

	t.Run("ChecksumValidation", func(t *testing.T) {
		// Create backup
		backupID, err := backupManager.CreateBackup()
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		// Get original backup info
		originalInfo, err := backupManager.GetBackupInfo(backupID)
		if err != nil {
			t.Fatalf("Failed to get original backup info: %v", err)
		}

		// Modify backup file to corrupt it
		backupPath := filepath.Join(backupManager.backupDir, backupID+".json")
		backupData, err := os.ReadFile(backupPath)
		if err != nil {
			t.Fatalf("Failed to read backup file: %v", err)
		}

		// Append some data to corrupt the file
		corruptedData := append(backupData, []byte(" corrupted")...)
		if err := os.WriteFile(backupPath, corruptedData, 0644); err != nil {
			t.Fatalf("Failed to write corrupted backup: %v", err)
		}

		// Get new backup info
		newInfo, err := backupManager.GetBackupInfo(backupID)
		if err != nil {
			t.Fatalf("Failed to get new backup info: %v", err)
		}

		// Checksums should be different
		if newInfo.Checksum == originalInfo.Checksum {
			t.Error("Checksums should be different after file corruption")
		}

		// Validation should detect the corruption if metadata exists
		metadata := backupManager.loadBackupMetadata(backupID)
		if metadata != nil && metadata.Checksum != "" {
			if err := backupManager.ValidateBackup(backupID); err == nil {
				t.Error("Validation should fail for corrupted backup with checksum mismatch")
			}
		}
	})
}

func TestBackupManagerErrors(t *testing.T) {
	// Test with non-existent config file
	nonExistentPath := "/nonexistent/path/config.json"
	backupManager := NewBackupManager(nonExistentPath)

	t.Run("CreateBackupNonExistentConfig", func(t *testing.T) {
		_, err := backupManager.CreateBackup()
		if err == nil {
			t.Error("Expected error when creating backup of non-existent config")
		}
	})

	t.Run("RestoreNonExistentBackup", func(t *testing.T) {
		err := backupManager.RestoreBackup("non_existent_backup")
		if err == nil {
			t.Error("Expected error when restoring non-existent backup")
		}
	})

	t.Run("DeleteNonExistentBackup", func(t *testing.T) {
		err := backupManager.DeleteBackup("non_existent_backup")
		if err == nil {
			t.Error("Expected error when deleting non-existent backup")
		}
	})

	t.Run("GetInfoNonExistentBackup", func(t *testing.T) {
		_, err := backupManager.GetBackupInfo("non_existent_backup")
		if err == nil {
			t.Error("Expected error when getting info for non-existent backup")
		}
	})

	t.Run("ExportNonExistentBackup", func(t *testing.T) {
		tempDir := t.TempDir()
		exportPath := filepath.Join(tempDir, "export.json")
		
		err := backupManager.ExportBackup("non_existent_backup", exportPath)
		if err == nil {
			t.Error("Expected error when exporting non-existent backup")
		}
	})

	t.Run("ImportInvalidConfig", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidPath := filepath.Join(tempDir, "invalid.json")
		
		// Create invalid JSON file
		if err := os.WriteFile(invalidPath, []byte("invalid json"), 0644); err != nil {
			t.Fatalf("Failed to create invalid file: %v", err)
		}

		_, err := backupManager.ImportBackup(invalidPath, "description")
		if err == nil {
			t.Error("Expected error when importing invalid config")
		}
	})
}