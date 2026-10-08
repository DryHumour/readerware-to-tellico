package extract

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:generate go run gen.go

var (
	// dbExts are the extensions of the files expected in the Readerware database directory.
	dbExts = [...]string{".data", ".properties", ".script"}
	//go:embed imagedumper.jar
	imageDumperJar []byte

	// ErrJavaNotFound indicates that no usable Java executable could be located.
	ErrJavaNotFound = errors.New("java executable not found")
)

// Images extracts images from a Readerware database using the ImageDumper tool.
// It takes a context for cancellation, the source database path, the destination directory,
// and an optional path to the Java executable. If javaPath is empty, it will attempt to find
// the Java executable in the system PATH or other common locations.
func Images(ctx context.Context, src, dst, javaPath string) error {
	javaExec, err := findJava(javaPath)
	if err != nil {
		return err // already wrapped with ErrJavaNotFound
	}

	dbStem, err := resolveDBPath(src)
	if err != nil {
		return err
	}

	if err := validateOutputPath(dst); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "readerware-image-extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	jarPath, err := setup(tmpDir)
	if err != nil {
		return err
	}

	cmd := command(ctx, javaExec, jarPath, dbStem, dst)
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("failed to run Readerware image extraction: %w: %w", ErrJavaNotFound, err)
		}
		return fmt.Errorf("failed to run Readerware image extraction using %q: %w", javaExec, err)
	}

	return nil
}

// findJava resolves the path to the Java executable.
func findJava(customPath string) (string, error) {
	// 1. Explicit Custom Path
	if customPath != "" {
		path, err := findJavaCustomPath(customPath)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrJavaNotFound, err)
		}
		return path, nil
	}

	// 2. Windows Registry Lookup (no-op on non-Windows)
	if installDir := findReaderwareInstallDir(); installDir != "" {
		target := filepath.Join(installDir, "jre", "bin", javaProg)
		if _, err := os.Stat(target); err == nil {
			return target, nil
		}
	}

	// 3. JAVA_HOME Environment Variable
	if jh := os.Getenv("JAVA_HOME"); jh != "" {
		target := filepath.Join(jh, "bin", javaProg)
		if _, err := os.Stat(target); err == nil {
			return target, nil
		}
	}

	// 4. JRE_HOME Environment Variable
	if jh := os.Getenv("JRE_HOME"); jh != "" {
		target := filepath.Join(jh, "bin", javaProg)
		if _, err := os.Stat(target); err == nil {
			return target, nil
		}
	}

	// 5. System PATH via exec.LookPath
	if p, err := exec.LookPath(javaProg); err == nil {
		return p, nil
	}

	// 6. Default Fallback
	return javaProg, nil
}

// resolveDBPath resolves the HSQLDB database file stem.
// If src is a directory, it looks for an HSQLDB stem named after the directory inside it.
// If src is a file, it trims off the HSQLDB suffix (such as .data, .properties, or .script)
// to get the pure file stem.
// It also verifies that the required database properties or script files exist before returning.
func resolveDBPath(src string) (string, error) {
	fi, err := os.Stat(src)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("database path does not exist: %s", src)
		}
		return "", fmt.Errorf("failed to read database path %q: %w", src, err)
	}

	stem := src
	if fi.IsDir() {
		dirName := filepath.Base(src)
		stem = filepath.Join(src, dirName)
	} else if stemExt := filepath.Ext(src); stemExt != "" {
		// Trim standard HSQLDB suffixes to find the base stem
		for _, ext := range dbExts {
			if strings.EqualFold(stemExt, ext) {
				stem = stem[:len(stem)-len(stemExt)]
				break
			}
		}
	}

	// Test that the standard HSQLDB files are present
	for _, ext := range dbExts {
		path := stem + ext
		if fi, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("cannot access required HSQLDB database file: %w", err)
		} else if fi.IsDir() {
			return "", fmt.Errorf("expected HSQLDB database file but found directory: %s", path)
		}
	}

	return stem, nil
}

// validateOutputPath verifies that the destination path is either non-existent or a valid directory.
func validateOutputPath(dst string) error {
	switch fi, err := os.Stat(dst); {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return fmt.Errorf("failed to read output path %q: %w", dst, err)
	case !fi.IsDir():
		return fmt.Errorf("output path exists but is not a directory: %s", dst)
	default:
		return nil
	}
}

// setup writes the embedded self-contained imagedumper.jar to the specified directory.
func setup(dir string) (string, error) {
	jarPath := filepath.Join(dir, "imagedumper.jar")
	if err := os.WriteFile(jarPath, imageDumperJar, 0o644); err != nil {
		return "", fmt.Errorf("failed to write image dumper jar: %w", err)
	}
	return jarPath, nil
}

// command creates an exec.Cmd for running the ImageDumper with the given parameters.
func command(ctx context.Context, javaExec, jarPath, src, dst string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, javaExec, "-jar", jarPath, src, dst)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil // allow java in current directory
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
