//go:build !windows

package extract

import (
	"os"
	"path/filepath"
)

const javaProg = "java"

// findJavaCustomPath resolves the full path to the java executable.
// If the path is a directory, it appends the java program name to it.
func findJavaCustomPath(path string) (string, error) {
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		path = filepath.Join(path, javaProg)
		fi, err = os.Stat(path)
	}
	if err != nil {
		return "", err
	}
	return path, nil
}

// findReaderwareInstallDir returns an empty string on non-Windows platforms.
func findReaderwareInstallDir() string {
	return ""
}
