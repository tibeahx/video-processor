package main

import (
	"errors"
	"fmt"
	"os"
)

func prepare(folders []string) error {
	if folders == nil {
		return errors.New("nil folders")
	}

	for _, folder := range folders {
		if err := os.MkdirAll(folder, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", folder, err)
		}
	}

	return nil
}

func cleanUp(dirsToClean []string) error {
	for _, dir := range dirsToClean {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to remove %s: %w", dir, err)
		}
	}

	return nil
}
