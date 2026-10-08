package extract

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
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

	stderr := newTailBuffer(4 << 10)
	cmd := command(ctx, javaExec, jarPath, dbStem, dst)
	cmd.Stderr = io.MultiWriter(os.Stderr, stderr)
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("failed to run Readerware image extraction using %q: %w", javaExec, err)
		if tail := strings.TrimSpace(stderr.String()); tail != "" {
			err = fmt.Errorf("%w\njava stderr: %s", err, tail)
		}
		return err
	}

	return nil
}

// findJava resolves the path to the Java executable, returning a validated
// absolute path or an error wrapping ErrJavaNotFound.
func findJava(customPath string) (string, error) {
	// 1. Explicit Custom Path
	if customPath != "" {
		path, err := findJavaCustomPath(customPath)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrJavaNotFound, err)
		}
		return filepath.Abs(path)
	}

	// 2. Windows Registry Lookup (no-op on non-Windows)
	if installDir := findReaderwareInstallDir(); installDir != "" {
		target := filepath.Join(installDir, "jre", "bin", javaProg)
		if _, err := os.Stat(target); err == nil {
			return filepath.Abs(target)
		}
	}

	// 3. JAVA_HOME / JRE_HOME Environment Variables
	for _, env := range []string{"JAVA_HOME", "JRE_HOME"} {
		if home := os.Getenv(env); home != "" {
			target := filepath.Join(home, "bin", javaProg)
			if _, err := os.Stat(target); err == nil {
				return filepath.Abs(target)
			}
		}
	}

	// 4. System PATH via exec.LookPath
	if p, err := exec.LookPath(javaProg); err == nil {
		return filepath.Abs(p)
	}

	return "", fmt.Errorf("%w: not found in Readerware installation, JAVA_HOME, JRE_HOME, or PATH", ErrJavaNotFound)
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

// tailBuffer is an io.Writer that retains only the last max bytes written,
// for surfacing subprocess output in error messages after a failure.
type tailBuffer struct {
	buf []byte
	max int
}

func newTailBuffer(max int) *tailBuffer {
	return &tailBuffer{max: max}
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	t.buf = append(t.buf, p...)
	if len(t.buf) > t.max {
		t.buf = append([]byte(nil), t.buf[len(t.buf)-t.max:]...)
	}
	return len(p), nil
}

func (t *tailBuffer) String() string {
	return string(t.buf)
}

// command creates an exec.Cmd for running the ImageDumper with the given parameters.
func command(ctx context.Context, javaExec, jarPath, src, dst string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, javaExec, "-jar", jarPath, src, dst)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
