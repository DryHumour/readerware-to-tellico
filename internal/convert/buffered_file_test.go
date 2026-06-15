package convert

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBufferedFile(t *testing.T) {
	t.Parallel()

	t.Run("file without BOM", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := []byte("test content")
		require.NoError(t, os.WriteFile(testFile, content, 0644))

		rc, err := newBufferedFile(testFile)
		require.NoError(t, err)
		require.NotNil(t, rc)

		// Read and verify content
		buf := make([]byte, len(content))
		n, err := rc.Read(buf)
		require.NoError(t, err)
		require.Equal(t, len(content), n)
		require.Equal(t, content, buf)

		require.NoError(t, rc.Close())
	})

	t.Run("file with UTF-8 BOM", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		bom := []byte{0xEF, 0xBB, 0xBF}
		content := []byte("test content")
		fullContent := append(bom, content...)
		require.NoError(t, os.WriteFile(testFile, fullContent, 0644))

		rc, err := newBufferedFile(testFile)
		require.NoError(t, err)
		require.NotNil(t, rc)

		// Read and verify BOM was stripped
		buf := make([]byte, len(content))
		n, err := rc.Read(buf)
		require.NoError(t, err)
		require.Equal(t, len(content), n)
		require.Equal(t, content, buf)

		require.NoError(t, rc.Close())
	})

	t.Run("file doesn't exist", func(t *testing.T) {
		t.Parallel()

		_, err := newBufferedFile("/nonexistent/file.txt")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to open file")
	})
}

func TestBufferedFile_Read(t *testing.T) {
	t.Parallel()

	t.Run("reads content correctly", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := []byte("hello world")
		require.NoError(t, os.WriteFile(testFile, content, 0644))

		rc, err := newBufferedFile(testFile)
		require.NoError(t, err)
		defer rc.Close()

		buf := make([]byte, 20)
		n, err := rc.Read(buf)
		require.NoError(t, err)
		require.Equal(t, len(content), n)
		require.Equal(t, content, buf[:n])
	})
}

func TestBufferedFile_Close(t *testing.T) {
	t.Parallel()

	t.Run("closes underlying file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := []byte("test")
		require.NoError(t, os.WriteFile(testFile, content, 0644))

		rc, err := newBufferedFile(testFile)
		require.NoError(t, err)

		require.NoError(t, rc.Close())

		// Closing again returns error (file already closed)
		require.Error(t, rc.Close())
	})
}
