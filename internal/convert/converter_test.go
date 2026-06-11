package convert

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/collection"
	"github.com/stretchr/testify/require"
)

var (
	testLogger = slog.Default()
)

func TestNewConverter(t *testing.T) {
	t.Parallel()

	t.Run("success with valid inputs", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := config.DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		cfg, err := config.New(dto, nil)
		require.NoError(t, err)
		defer cfg.Close()

		policy := collection.NewBooksPolicy(t.Context(), cfg)
		converter, err := NewConverter(t.Context(), cfg, policy)
		require.NoError(t, err)
		require.NotNil(t, converter)
	})

	t.Run("error when policy is nil", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := filepath.Join(tmpDir, "output.tc")

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := config.DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		cfg, err := config.New(dto, nil)
		require.NoError(t, err)
		defer cfg.Close()

		_, err = NewConverter(t.Context(), cfg, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing policy")
	})

	t.Run("error when reader is nil", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		outputFile := filepath.Join(tmpDir, "output.tc")

		dto := config.DTO{
			InputFile:  filepath.Join(tmpDir, "nonexistent.csv"),
			OutputFile: outputFile,
		}

		_, err := config.New(dto, nil)
		require.Error(t, err)
	})

	t.Run("error when writer is nil", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := "/invalid/path/output.tc"

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		dto := config.DTO{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		_, err := config.New(dto, nil)
		require.Error(t, err)
	})
}

func TestConverter_Run_EmptyInput(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "empty.csv")
	outputFile := filepath.Join(tmpDir, "output.tc")

	// Create empty CSV file
	require.NoError(t, os.WriteFile(inputFile, []byte(""), 0644))

	dto := config.DTO{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}

	// Empty file causes error at config level (BOM peek fails)
	cfg, err := config.New(dto, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "EOF")
	_ = cfg
}

func TestConverter_Run_ContextCancellation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.csv")
	outputFile := filepath.Join(tmpDir, "output.tc")

	// Create CSV with headers
	csvContent := `ROW#,TITLE,AUTHOR
1,Test Book,John Doe
2,Another Book,Jane Smith
3,Third Book,Bob Johnson`
	require.NoError(t, os.WriteFile(inputFile, []byte(csvContent), 0644))

	dto := config.DTO{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}

	cfg, err := config.New(dto, nil)
	require.NoError(t, err)
	defer cfg.Close()

	policy := collection.NewBooksPolicy(t.Context(), cfg)
	converter, err := NewConverter(t.Context(), cfg, policy)
	require.NoError(t, err)

	// Cancel context immediately
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	errSeen := false
	for _, err := range converter.Run(ctx) {
		if err != nil {
			errSeen = true
			// Context cancellation is an error
			require.Error(t, err)
		}
	}
	require.True(t, errSeen, "context cancellation should produce at least one error")
}

func TestConverter_Run_ValidCSV(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.csv")
	outputFile := filepath.Join(tmpDir, "output.tc")

	// Create valid CSV
	csvContent := `ROW#,TITLE,AUTHOR
1,Test Book,John Doe
2,Another Book,Jane Smith`
	require.NoError(t, os.WriteFile(inputFile, []byte(csvContent), 0644))

	dto := config.DTO{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}

	cfg, err := config.New(dto, testLogger)
	require.NoError(t, err)
	defer cfg.Close()

	policy := collection.NewBooksPolicy(t.Context(), cfg)
	converter, err := NewConverter(t.Context(), cfg, policy)
	require.NoError(t, err)

	for report, err := range converter.Run(t.Context()) {
		require.NoError(t, err)
		_ = report
	}

	// Verify output file was created
	info, err := os.Stat(outputFile)
	require.NoError(t, err)
	require.NotZero(t, info.Size(), "output file should not be empty")
}

func TestConverter_Run_InvalidCSV(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "invalid.csv")
	outputFile := filepath.Join(tmpDir, "output.tc")

	// Create malformed CSV (unclosed quote)
	csvContent := `ROW#,TITLE,AUTHOR
1,"Test Book,John Doe`
	require.NoError(t, os.WriteFile(inputFile, []byte(csvContent), 0644))

	dto := config.DTO{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}

	cfg, err := config.New(dto, testLogger)
	require.NoError(t, err)
	defer cfg.Close()

	policy := collection.NewBooksPolicy(t.Context(), cfg)
	converter, err := NewConverter(t.Context(), cfg, policy)
	require.NoError(t, err)

	errSeen := false
	for _, err := range converter.Run(t.Context()) {
		if err != nil {
			errSeen = true
			require.Error(t, err)
		}
	}
	// CSV parsing error should be reported
	require.True(t, errSeen, "expected error for malformed CSV")
}

func TestConverter_Run_OnlyOnce(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.csv")
	outputFile := filepath.Join(tmpDir, "output.tc")

	// Create valid CSV
	csvContent := `ROW#,TITLE,AUTHOR
1,Test Book,John Doe`
	require.NoError(t, os.WriteFile(inputFile, []byte(csvContent), 0644))

	dto := config.DTO{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}

	cfg, err := config.New(dto, testLogger)
	require.NoError(t, err)
	defer cfg.Close()

	policy := collection.NewBooksPolicy(t.Context(), cfg)
	converter, err := NewConverter(t.Context(), cfg, policy)
	require.NoError(t, err)

	// First run should succeed
	for report, err := range converter.Run(t.Context()) {
		require.NoError(t, err)
		_ = report
	}

	// Second run should be a no-op (idempotent)
	reportCount := 0
	for report, err := range converter.Run(t.Context()) {
		require.NoError(t, err)
		_ = report
		reportCount++
	}
	require.Equal(t, 0, reportCount, "second run should produce no reports")
}
