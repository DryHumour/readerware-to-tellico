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

	"golang.org/x/sync/errgroup"
)

//go:generate go run gen.go

var (
	// dbExts are the extensions of the files expected in the Readerware database directory.
	dbExts = [...]string{".data", ".properties", ".script"}
	//go:embed ImageDumper.class
	imageDumperClass []byte
	//go:embed hsqldb.jar
	hsqldbJar []byte
)

// Images extracts images from a Readerware database using the ImageDumper tool.
// It takes a context for cancellation, the source database path, the destination directory,
// and an optional path to the Java executable. If javaPath is empty, it will attempt to find
// the Java executable in the system PATH or other common locations.
func Images(ctx context.Context, src, dst, javaPath string) error {
	javaExec, err := findJava(javaPath)
	if err != nil {
		return fmt.Errorf("failed to locate Java executable: %w", err)
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

	if err := setup(tmpDir); err != nil {
		return err
	}

	cmd := command(ctx, javaExec, tmpDir, dbStem, dst)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run Readerware image extraction using %q: %w", javaExec, err)
	}

	return nil
}

// findJava resolves the path to the Java executable.
func findJava(customPath string) (string, error) {
	// 1. Explicit Custom Path
	if customPath != "" {
		return findJavaCustomPath(customPath)
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

// setup writes the embedded ImageDumper.class and hsqldb.jar files to the specified directory.
func setup(dir string) error {
	var g errgroup.Group
	g.Go(func() error {
		return os.WriteFile(filepath.Join(dir, "ImageDumper.class"), imageDumperClass, 0o644)
	})
	g.Go(func() error {
		return os.WriteFile(filepath.Join(dir, "hsqldb.jar"), hsqldbJar, 0o644)
	})
	return g.Wait()
}

// command creates an exec.Cmd for running the ImageDumper with the given parameters.
func command(ctx context.Context, javaExec, path, src, dst string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, javaExec, "ImageDumper", src, dst)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil // allow java in current directory
	}
	parts := []string{filepath.Join(path, "hsqldb.jar"), path}
	if classpath := os.Getenv("CLASSPATH"); classpath != "" {
		// honour user's CLASSPATH, but add our dependencies
		parts = append([]string{classpath}, parts...)
	}
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "CLASSPATH="+strings.Join(parts, string(filepath.ListSeparator)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
