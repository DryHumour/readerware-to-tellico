package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDTO_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid DTO with all fields", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "test.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")
		imagesDir := filepath.Join(tmpDir, "images")

		require.NoError(t, os.Mkdir(imagesDir, 0755))
		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
			ImagesDirs: Directories{First: imagesDir},
			Level:      "info",
		}

		err := dto.Validate()
		require.NoError(t, err)
	})

	t.Run("valid DTO with optional fields omitted", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "test.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		err := dto.Validate()
		require.NoError(t, err)
	})

	t.Run("missing input-file", func(t *testing.T) {
		t.Parallel()

		dto := DTO{
			OutputFile: "output.tc",
		}

		err := dto.Validate()
		require.Error(t, err)
	})

	t.Run("missing output-file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "test.csv")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile: inputFile,
		}

		err := dto.Validate()
		require.Error(t, err)
	})

	t.Run("invalid log-level", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "test.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
			Level:      "invalid",
		}

		err := dto.Validate()
		require.Error(t, err)
	})

	t.Run("valid log-level - debug", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "test.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
			Level:      "debug",
		}

		err := dto.Validate()
		require.NoError(t, err)
	})

	t.Run("valid log-level - warn", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "test.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
			Level:      "warn",
		}

		err := dto.Validate()
		require.NoError(t, err)
	})

	t.Run("valid log-level - error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "test.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
			Level:      "error",
		}

		err := dto.Validate()
		require.NoError(t, err)
	})
}
