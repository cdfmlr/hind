package cgroups

import (
	"fmt"
	"os"
)

// echo $line >> $filePath
func appendFile(filePath string, line string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0666)
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
