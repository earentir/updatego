package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// AllocateBackupDir returns parent/baseName or parent/baseName-N when baseName exists.
func AllocateBackupDir(parent, baseName string) (string, error) {
	candidate := filepath.Join(parent, baseName)
	if !IsDirExists(candidate) {
		return candidate, nil
	}
	for i := 2; ; i++ {
		candidate = filepath.Join(parent, fmt.Sprintf("%s-%d", baseName, i))
		if !IsDirExists(candidate) {
			return candidate, nil
		}
	}
}

// BackupOldGo moves goFullPath to backupPath without deleting an existing backup directory.
func BackupOldGo(backupPath, goFullPath string) error {
	if !IsDirExists(goFullPath) {
		return nil
	}
	dest := backupPath
	if IsDirExists(dest) {
		alt, err := AllocateBackupDir(filepath.Dir(backupPath), filepath.Base(backupPath))
		if err != nil {
			return err
		}
		dest = alt
	}
	if err := os.Rename(goFullPath, dest); err != nil {
		return fmt.Errorf("error renaming the old Go folder to %s: %w", dest, err)
	}
	return nil
}
