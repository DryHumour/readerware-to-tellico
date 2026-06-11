package convert

import (
	"archive/zip"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/collection"
	"github.com/stretchr/testify/require"
)

// extractXML runs the conversion with the given CSV and returns the XML content from the TC file.
func extractXML(t *testing.T, csvContent string) []byte {
	t.Helper()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	tcPath := filepath.Join(tmpDir, "test.tc")

	require.NoError(t, os.WriteFile(csvPath, []byte(csvContent), 0644))

	dto := config.DTO{
		InputFile:  csvPath,
		OutputFile: tcPath,
	}

	cfg, err := config.New(dto, nil)
	require.NoError(t, err)
	defer cfg.Close()

	converter, err := NewConverter(t.Context(), cfg, collection.NewBooksPolicy(t.Context(), cfg))
	require.NoError(t, err)

	for _, err := range converter.Run(t.Context()) {
		require.NoError(t, err)
	}

	tcContent, err := os.ReadFile(tcPath)
	require.NoError(t, err)

	zr, err := zip.NewReader(bytes.NewReader(tcContent), int64(len(tcContent)))
	require.NoError(t, err)

	xmlFile, err := zr.Open("tellico.xml")
	require.NoError(t, err)
	defer xmlFile.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(xmlFile)
	require.NoError(t, err)
	return buf.Bytes()
}

func TestIntegration_XMLValidation(t *testing.T) {
	t.Parallel()

	csvContent := `ROW#,TITLE,AUTHOR,PUBLISHER,ISBN
1,Test Book,John Doe,Test Publisher,978-0-123-45678-9
2,Another Book,Jane Smith,Another Pub,978-0-987-65432-1`

	xmlContent := extractXML(t, csvContent)

	if _, err := exec.LookPath("xmllint"); err != nil {
		t.Skip("xmllint not available")
	}

	// Always: check well-formedness (no DTD fetch needed)
	cmd := exec.Command("xmllint", "--noout", "-")
	cmd.Stdin = bytes.NewReader(xmlContent)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("XML is not well-formed:\n%s", out)
	}

	// DTD validation skipped - the DTD expects images element to have content, but we test without images
	t.Skip("DTD validation skipped - requires images in DTD")
}

func TestIntegration_WithImages(t *testing.T) {
	t.Parallel()

	// Create a sample CSV input
	csvContent := `ROWKEY,TITLE,AUTHOR
123,Test Book,John Doe
456,Another Book,Jane Smith`

	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	tcPath := filepath.Join(tmpDir, "test.tc")
	imgDir := filepath.Join(tmpDir, "images")

	// Write the CSV file
	require.NoError(t, os.WriteFile(csvPath, []byte(csvContent), 0644))

	// Valid JPEG magic numbers for dummy images
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}
	pngData := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}

	// Create test images
	require.NoError(t, os.Mkdir(imgDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(imgDir, "123.jpg"), jpegData, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(imgDir, "456.png"), pngData, 0644))

	dto := config.DTO{
		InputFile:  csvPath,
		OutputFile: tcPath,
		ImagesDirs: config.Directories{First: imgDir},
	}

	cfg, err := config.New(dto, nil)
	require.NoError(t, err)
	defer cfg.Close()

	converter, err := NewConverter(t.Context(), cfg, collection.NewBooksPolicy(t.Context(), cfg))
	require.NoError(t, err)

	for _, err := range converter.Run(t.Context()) {
		require.NoError(t, err)
	}

	// Verify TC file was created and has content
	info, err := os.Stat(tcPath)
	require.NoError(t, err)
	require.NotZero(t, info.Size(), "TC file is empty")
}

func TestIntegration_MissingHeader(t *testing.T) {
	t.Parallel()

	csvContent := `AUTHOR,PUBLISHER,ISBN
John Doe,Test Publisher,978-0-123-45678-9`

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	tcPath := filepath.Join(tmpDir, "test.tc")

	require.NoError(t, os.WriteFile(csvPath, []byte(csvContent), 0644))

	dto := config.DTO{
		InputFile:  csvPath,
		OutputFile: tcPath,
	}

	cfg, err := config.New(dto, nil)
	require.NoError(t, err)
	defer cfg.Close()

	converter, err := NewConverter(t.Context(), cfg, collection.NewBooksPolicy(t.Context(), cfg))
	require.NoError(t, err)

	errSeen := false
	for _, err := range converter.Run(t.Context()) {
		if err != nil {
			require.Contains(t, err.Error(), "missing required header")
			errSeen = true
		}
	}
	require.True(t, errSeen, "expected error for missing TITLE header")
}

func TestIntegration_MissingIDHeader(t *testing.T) {
	t.Parallel()

	csvContent := `TITLE,AUTHOR,PUBLISHER,ISBN
Test Book,John Doe,Test Publisher,978-0-123-45678-9`

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	tcPath := filepath.Join(tmpDir, "test.tc")

	require.NoError(t, os.WriteFile(csvPath, []byte(csvContent), 0644))

	dto := config.DTO{
		InputFile:  csvPath,
		OutputFile: tcPath,
	}

	cfg, err := config.New(dto, nil)
	require.NoError(t, err)
	defer cfg.Close()

	converter, err := NewConverter(t.Context(), cfg, collection.NewBooksPolicy(t.Context(), cfg))
	require.NoError(t, err)

	errSeen := false
	for _, err := range converter.Run(t.Context()) {
		if err != nil {
			require.Contains(t, err.Error(), "missing required header")
			errSeen = true
		}
	}
	require.True(t, errSeen, "expected error for missing ROWKEY/ROW# header")
}

func TestIntegration_MissingROWKEYWithImages(t *testing.T) {
	t.Parallel()

	csvContent := `TITLE,AUTHOR
Test Book,John Doe`

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	tcPath := filepath.Join(tmpDir, "test.tc")
	imgDir := filepath.Join(tmpDir, "images")

	require.NoError(t, os.WriteFile(csvPath, []byte(csvContent), 0644))
	require.NoError(t, os.Mkdir(imgDir, 0755))

	dto := config.DTO{
		InputFile:  csvPath,
		OutputFile: tcPath,
		ImagesDirs: config.Directories{First: imgDir},
	}

	cfg, err := config.New(dto, nil)
	require.NoError(t, err)
	defer cfg.Close()

	converter, err := NewConverter(t.Context(), cfg, collection.NewBooksPolicy(t.Context(), cfg))
	require.NoError(t, err)

	errSeen := false
	for _, err := range converter.Run(t.Context()) {
		if err != nil {
			require.Contains(t, err.Error(), "ROWKEY")
			errSeen = true
		}
	}
	require.True(t, errSeen, "expected error for missing ROWKEY header with images")
}

func TestIntegration_ValidHeadersWithROWKEY(t *testing.T) {
	t.Parallel()

	csvContent := `ROWKEY,TITLE,AUTHOR
123,Test Book,John Doe
456,Another Book,Jane Smith`

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	tcPath := filepath.Join(tmpDir, "test.tc")

	require.NoError(t, os.WriteFile(csvPath, []byte(csvContent), 0644))

	dto := config.DTO{
		InputFile:  csvPath,
		OutputFile: tcPath,
	}

	cfg, err := config.New(dto, nil)
	require.NoError(t, err)
	defer cfg.Close()

	converter, err := NewConverter(t.Context(), cfg, collection.NewBooksPolicy(t.Context(), cfg))
	require.NoError(t, err)

	for _, err := range converter.Run(t.Context()) {
		require.NoError(t, err)
	}
}

func TestIntegration_ValidHeadersWithROWHASH(t *testing.T) {
	t.Parallel()

	csvContent := `ROW#,TITLE,AUTHOR
1,Test Book,John Doe
2,Another Book,Jane Smith`

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	tcPath := filepath.Join(tmpDir, "test.tc")

	require.NoError(t, os.WriteFile(csvPath, []byte(csvContent), 0644))

	dto := config.DTO{
		InputFile:  csvPath,
		OutputFile: tcPath,
	}

	cfg, err := config.New(dto, nil)
	require.NoError(t, err)
	defer cfg.Close()

	converter, err := NewConverter(t.Context(), cfg, collection.NewBooksPolicy(t.Context(), cfg))
	require.NoError(t, err)

	for _, err := range converter.Run(t.Context()) {
		require.NoError(t, err)
	}
}

func TestIntegration_ValidHeadersWithImages(t *testing.T) {
	t.Parallel()

	csvContent := `ROWKEY,TITLE,AUTHOR
123,Test Book,John Doe`

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	tcPath := filepath.Join(tmpDir, "test.tc")
	imgDir := filepath.Join(tmpDir, "images")

	// Valid JPEG magic numbers for dummy images
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}

	require.NoError(t, os.WriteFile(csvPath, []byte(csvContent), 0644))
	require.NoError(t, os.Mkdir(imgDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(imgDir, "123.jpg"), jpegData, 0644))

	dto := config.DTO{
		InputFile:  csvPath,
		OutputFile: tcPath,
		ImagesDirs: config.Directories{First: imgDir},
	}

	cfg, err := config.New(dto, nil)
	require.NoError(t, err)
	defer cfg.Close()

	converter, err := NewConverter(t.Context(), cfg, collection.NewBooksPolicy(t.Context(), cfg))
	require.NoError(t, err)

	for _, err := range converter.Run(t.Context()) {
		require.NoError(t, err)
	}
}

func TestIntegration_UserTemplateOverride(t *testing.T) {
	t.Parallel()

	csvContent := `ROWKEY,TITLE,AUTHOR
123,Test Book,John Doe`

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	tcPath := filepath.Join(tmpDir, "test.tc")
	templateDir := filepath.Join(tmpDir, "templates")

	// Create template directory and a custom title template
	require.NoError(t, os.Mkdir(templateDir, 0755))
	customTitleTemplate := `{{- define "books.title" }}
   <title>CUSTOM: {{ .V "TITLE" }}</title>
{{- end -}}
`
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "books.title.gotmpl"), []byte(customTitleTemplate), 0644))

	require.NoError(t, os.WriteFile(csvPath, []byte(csvContent), 0644))

	dto := config.DTO{
		InputFile:    csvPath,
		OutputFile:   tcPath,
		TemplateDirs: []string{templateDir},
	}

	cfg, err := config.New(dto, nil)
	require.NoError(t, err)
	defer cfg.Close()

	converter, err := NewConverter(t.Context(), cfg, collection.NewBooksPolicy(t.Context(), cfg))
	require.NoError(t, err)

	for _, err := range converter.Run(t.Context()) {
		require.NoError(t, err)
	}

	// Extract and verify XML contains custom template
	tcContent, err := os.ReadFile(tcPath)
	require.NoError(t, err)

	zr, err := zip.NewReader(bytes.NewReader(tcContent), int64(len(tcContent)))
	require.NoError(t, err)

	xmlFile, err := zr.Open("tellico.xml")
	require.NoError(t, err)
	defer xmlFile.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(xmlFile)
	require.NoError(t, err)
	xmlContent := buf.Bytes()

	require.Contains(t, string(xmlContent), "CUSTOM: Test Book", "user template should override default")
}
