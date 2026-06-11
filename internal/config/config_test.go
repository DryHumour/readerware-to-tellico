package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_New(t *testing.T) {
	t.Parallel()

	t.Run("success with valid input and output files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		// Create input file
		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		cfg, err := New(dto, nil)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.NotNil(t, cfg.Reader)
		require.NotNil(t, cfg.Writer)
		require.Nil(t, cfg.Logger) // nil logger is allowed
		require.Empty(t, cfg.ImagesDirs)

		require.NoError(t, cfg.Close())
	})

	t.Run("success with images directories", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")
		imagesDir := filepath.Join(tmpDir, "images")

		// Create input file and images directory
		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))
		require.NoError(t, os.Mkdir(imagesDir, 0755))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
			ImagesDirs: Directories{First: imagesDir},
		}

		cfg, err := New(dto, nil)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, dto.ImagesDirs, cfg.ImagesDirs)

		require.NoError(t, cfg.Close())
	})

	t.Run("success with template directories", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")
		templateDir1 := filepath.Join(tmpDir, "templates1")
		templateDir2 := filepath.Join(tmpDir, "templates2")

		// Create input file and template directories
		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))
		require.NoError(t, os.Mkdir(templateDir1, 0755))
		require.NoError(t, os.Mkdir(templateDir2, 0755))

		dto := DTO{
			InputFile:    inputFile,
			OutputFile:   outputFile,
			TemplateDirs: []string{templateDir1, templateDir2},
		}

		cfg, err := New(dto, nil)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, dto.TemplateDirs, cfg.TemplateDirs)

		require.NoError(t, cfg.Close())
	})

	t.Run("success with output in current directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := "output.tc" // No directory prefix

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		cfg, err := New(dto, nil)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		require.NoError(t, cfg.Close())
		os.Remove(outputFile) // Clean up
	})

	t.Run("error when input file doesn't exist", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		outputFile := filepath.Join(tmpDir, "output.tc")

		dto := DTO{
			InputFile:  filepath.Join(tmpDir, "nonexistent.csv"),
			OutputFile: outputFile,
		}

		_, err := New(dto, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to open file")
	})

	t.Run("error when output file creation fails", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		// Use a path that would require creating a directory in an invalid location
		outputFile := "/invalid/path/output.tc"

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		_, err := New(dto, nil)
		require.Error(t, err)
	})

	t.Run("custom logger is used", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		customLogger := slog.Default()

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		cfg, err := New(dto, customLogger)
		require.NoError(t, err)
		require.Same(t, customLogger, cfg.Logger)

		require.NoError(t, cfg.Close())
	})
}

func TestConfig_Close(t *testing.T) {
	t.Parallel()

	t.Run("success closing both reader and writer", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		cfg, err := New(dto, nil)
		require.NoError(t, err)

		err = cfg.Close()
		require.NoError(t, err)
	})

	t.Run("idempotent - can close multiple times", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		cfg, err := New(dto, nil)
		require.NoError(t, err)

		require.NoError(t, cfg.Close())
		// Second close is safe and returns nil (idempotent)
		require.NoError(t, cfg.Close())
	})

	t.Run("error when reader close fails", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		cfg, err := New(dto, nil)
		require.NoError(t, err)

		// Manually close reader to simulate failure
		cfg.Reader.Close()

		err = cfg.Close()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to close reader")
	})

	t.Run("error when writer close fails", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		cfg, err := New(dto, nil)
		require.NoError(t, err)

		// Manually close writer to simulate failure
		cfg.Writer.Close()

		err = cfg.Close()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to close writer")
	})
}

func TestConfig_ImageFS(t *testing.T) {
	t.Parallel()

	t.Run("empty when no images directories specified", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		cfg, err := New(dto, nil)
		require.NoError(t, err)
		require.Empty(t, cfg.ImagesDirs)

		require.NoError(t, cfg.Close())
	})

	t.Run("contains filesystems for each directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")
		imagesDir1 := filepath.Join(tmpDir, "images1")
		imagesDir2 := filepath.Join(tmpDir, "images2")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))
		require.NoError(t, os.Mkdir(imagesDir1, 0755))
		require.NoError(t, os.Mkdir(imagesDir2, 0755))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
			ImagesDirs: Directories{First: imagesDir1, Second: imagesDir2},
		}

		cfg, err := New(dto, nil)
		require.NoError(t, err)
		require.Equal(t, dto.ImagesDirs, cfg.ImagesDirs)

		require.NoError(t, cfg.Close())
	})
}
