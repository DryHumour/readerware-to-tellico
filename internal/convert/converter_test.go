package convert

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/collection"
	"github.com/stretchr/testify/assert"
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

		cfg := config.Config{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		policy := collection.NewBooksPolicy()
		converter := NewConverter(cfg, policy)
		require.NotNil(t, converter)
	})

	t.Run("Run fails when input file missing", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		outputFile := filepath.Join(tmpDir, "output.tc")

		cfg := config.Config{
			InputFile:  filepath.Join(tmpDir, "nonexistent.csv"),
			OutputFile: outputFile,
		}

		policy := collection.NewBooksPolicy()
		converter := NewConverter(cfg, policy)

		var runErr error
		for _, err := range converter.Run(t.Context()) {
			if err != nil {
				runErr = err
			}
		}
		assert.ErrorContains(t, runErr, "failed to open input file", "should fail when input file does not exist")
	})

	t.Run("Run fails when output path invalid", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.csv")
		outputFile := "/invalid/path/output.tc"

		require.NoError(t, os.WriteFile(inputFile, []byte("test"), 0644))

		cfg := config.Config{
			InputFile:  inputFile,
			OutputFile: outputFile,
		}

		policy := collection.NewBooksPolicy()
		converter := NewConverter(cfg, policy)

		var runErr error
		for _, err := range converter.Run(t.Context()) {
			if err != nil {
				runErr = err
			}
		}
		assert.Error(t, runErr, "should fail when output path is invalid")
	})
}

func TestConverter_Run_EmptyInput(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "empty.csv")
	outputFile := filepath.Join(tmpDir, "output.tc")

	// Create empty CSV file
	require.NoError(t, os.WriteFile(inputFile, []byte(""), 0644))

	cfg := config.Config{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}

	policy := collection.NewBooksPolicy()
	converter := NewConverter(cfg, policy)

	var runErr error
	for _, err := range converter.Run(t.Context()) {
		if err != nil {
			runErr = err
		}
	}
	// Empty file causes BOM peek to fail with EOF
	assert.ErrorContains(t, runErr, "EOF", "empty file should fail with EOF error")
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

	cfg := config.Config{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}

	policy := collection.NewBooksPolicy()
	converter := NewConverter(cfg, policy)

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

	cfg := config.Config{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}

	policy := collection.NewBooksPolicy()
	converter := NewConverter(cfg, policy, WithLogger(testLogger))

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

	cfg := config.Config{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}

	policy := collection.NewBooksPolicy()
	converter := NewConverter(cfg, policy, WithLogger(testLogger))

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

	cfg := config.Config{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}

	policy := collection.NewBooksPolicy()
	converter := NewConverter(cfg, policy, WithLogger(testLogger))

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
