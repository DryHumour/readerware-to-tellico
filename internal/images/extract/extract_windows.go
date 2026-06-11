//go:build windows

package extract

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const javaProg = "java.exe"

// findJavaCustomPath resolves the full path to the java executable.
// If the path is a directory, it appends the java program name to it.
func findJavaCustomPath(path string) (string, error) {
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		path = filepath.Join(path, javaProg)
		fi, err = os.Stat(path)
	} else if err != nil && !strings.EqualFold(filepath.Ext(path), ".exe") {
		path += ".exe"
		fi, err = os.Stat(path)
	}
	if err != nil {
		return "", err
	}
	return path, nil
}

// findReaderwareInstallDir queries the Windows Registry for the Readerware 4 installation path.
func findReaderwareInstallDir() string {
	keys := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Readerware 4`,
		`SOFTWARE\Wow6432Node\Microsoft\Windows\CurrentVersion\Uninstall\Readerware 4`,
	}

	for _, path := range keys {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		loc, _, err := k.GetStringValue("InstallLocation")
		k.Close()
		if err == nil && loc != "" {
			return loc
		}
	}
	return ""
}
