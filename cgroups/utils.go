package cgroups

// This file implements help funcitons to write/append file.

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// echo $line >> $filePath
func appendFile(filePath string, line string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the text to the file
	_, err = fmt.Fprintf(f, "%s\n", line)
	return err
}

// echo $content > $filePath
func overwriteFile(filePath string, content string) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) // xxx: use Create instead
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return err
	}

	return nil
}

// mkdir -p $dirPath
func mkdirIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}
	return nil
}

// assertFsType checks if the filesystem type of path is fsType.
//
//	test $(stat -fc %T $path) = $fsType
func assertFsType(path string, fsType string) error {
	// XXX: use os.Stat instead
	out, err := exec.Command("stat", "-fc", "%T", path).Output()
	if err != nil {
		return fmt.Errorf("error checking filesystem type: %w", err)
	}

	got := strings.TrimSpace(string(out))
	if got != fsType {
		return fmt.Errorf("unexpected %s filesystem: %s (expected %s)",
			got, path, fsType)
	}

	return nil
}
