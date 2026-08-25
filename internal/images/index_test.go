package images

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"github.com/stretchr/testify/require"
)

func TestBuildIndex(t *testing.T) {
	t.Parallel()

	t.Run("success with valid directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		imgDir := filepath.Join(tmpDir, "images")

		require.NoError(t, os.Mkdir(imgDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(imgDir, "123.jpg"), []byte("fake image 1"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(imgDir, "456.png"), []byte("fake image 2"), 0644))

		dirs := config.Directories{First: imgDir}
		index, err := BuildIndex(t.Context(), dirs)
		require.NoError(t, err)
		require.NotNil(t, index)
		require.False(t, index.IsEmpty())
	})

	t.Run("success with Images subdirectory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		parentDir := filepath.Join(tmpDir, "readerware")
		imagesDir := filepath.Join(parentDir, "Images")

		require.NoError(t, os.MkdirAll(imagesDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(imagesDir, "789.gif"), []byte("fake image"), 0644))

		dirs := config.Directories{First: parentDir}
		index, err := BuildIndex(t.Context(), dirs)
		require.NoError(t, err)
		require.NotNil(t, index)
		require.False(t, index.IsEmpty())
	})

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		imgDir := filepath.Join(tmpDir, "images")

		require.NoError(t, os.Mkdir(imgDir, 0755))

		dirs := config.Directories{First: imgDir}
		index, err := BuildIndex(t.Context(), dirs)
		require.NoError(t, err)
		require.NotNil(t, index)
		require.True(t, index.IsEmpty())
	})

	t.Run("error on duplicate ROWKEY formats", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		imgDir := filepath.Join(tmpDir, "images")

		require.NoError(t, os.Mkdir(imgDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(imgDir, "123.jpg"), []byte("fake image 1"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(imgDir, "123.png"), []byte("fake image 2"), 0644))

		dirs := config.Directories{First: imgDir}
		_, err := BuildIndex(t.Context(), dirs)
		require.Error(t, err)
		require.Contains(t, err.Error(), "multiple image formats")
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		imgDir := filepath.Join(tmpDir, "nonexistent")

		dirs := config.Directories{First: imgDir}
		_, err := BuildIndex(t.Context(), dirs)
		require.Error(t, err)
	})

	t.Run("multiple directories", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		imgDir1 := filepath.Join(tmpDir, "images1")
		imgDir2 := filepath.Join(tmpDir, "images2")

		require.NoError(t, os.Mkdir(imgDir1, 0755))
		require.NoError(t, os.Mkdir(imgDir2, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(imgDir1, "123.jpg"), []byte("fake image 1"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(imgDir2, "456.png"), []byte("fake image 2"), 0644))

		dirs := config.Directories{First: imgDir1, Second: imgDir2}
		index, err := BuildIndex(t.Context(), dirs)
		require.NoError(t, err)
		require.NotNil(t, index)
		require.False(t, index.IsEmpty())
	})
}

func TestIndex_Row(t *testing.T) {
	t.Parallel()

	t.Run("returns row with matching images", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		imgDir := filepath.Join(tmpDir, "images")

		require.NoError(t, os.Mkdir(imgDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(imgDir, "123.jpg"), []byte("fake image"), 0644))

		dirs := config.Directories{First: imgDir}
		index, err := BuildIndex(t.Context(), dirs)
		require.NoError(t, err)

		// Check if index is not empty
		require.False(t, index.IsEmpty(), "index should not be empty")

		// Try to lookup the row
		row, ok := index.Row("123")
		require.True(t, ok, "row lookup should succeed for existing rowkey")
		// The row should have the entry in one of the slots
		hasEntry := false
		for _, entry := range row {
			if entry != nil {
				hasEntry = true
				break
			}
		}
		require.True(t, hasEntry, "row should have at least one entry")
	})

	t.Run("returns empty row for nonexistent key", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		imgDir := filepath.Join(tmpDir, "images")

		require.NoError(t, os.Mkdir(imgDir, 0755))

		dirs := config.Directories{First: imgDir}
		index, err := BuildIndex(t.Context(), dirs)
		require.NoError(t, err)

		row, ok := index.Row("999")
		require.False(t, ok)
		require.Empty(t, row)
	})

	t.Run("handles nil index", func(t *testing.T) {
		t.Parallel()

		var index *Index
		row, ok := index.Row("123")
		require.False(t, ok)
		require.Empty(t, row)
	})
}

func TestIndex_AllUsed(t *testing.T) {
	t.Parallel()

	t.Run("iterates over used images", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		imgDir := filepath.Join(tmpDir, "images")

		require.NoError(t, os.Mkdir(imgDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(imgDir, "123.jpg"), []byte("fake image"), 0644))

		dirs := config.Directories{First: imgDir}
		index, err := BuildIndex(t.Context(), dirs)
		require.NoError(t, err)

		count := 0
		for entry := range index.AllUsed() {
			require.NotNil(t, entry)
			count++
		}
		// By default, images are not marked as used
		require.Equal(t, 0, count)
	})

	t.Run("handles nil index", func(t *testing.T) {
		t.Parallel()

		var index *Index
		count := 0
		for range index.AllUsed() {
			count++
		}
		require.Equal(t, 0, count)
	})
}

func TestIndex_IsEmpty(t *testing.T) {
	t.Parallel()

	t.Run("returns true for empty index", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		imgDir := filepath.Join(tmpDir, "images")

		require.NoError(t, os.Mkdir(imgDir, 0755))

		dirs := config.Directories{First: imgDir}
		index, err := BuildIndex(t.Context(), dirs)
		require.NoError(t, err)
		require.True(t, index.IsEmpty())
	})

	t.Run("returns false for non-empty index", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		imgDir := filepath.Join(tmpDir, "images")

		require.NoError(t, os.Mkdir(imgDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(imgDir, "123.jpg"), []byte("fake image"), 0644))

		dirs := config.Directories{First: imgDir}
		index, err := BuildIndex(t.Context(), dirs)
		require.NoError(t, err)
		require.False(t, index.IsEmpty())
	})

	t.Run("handles nil index", func(t *testing.T) {
		t.Parallel()

		var index *Index
		require.True(t, index.IsEmpty())
	})
}

func TestGlobImage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		filename string
		wantKey  string
		wantExt  string
		wantOk   bool
	}{
		{"valid jpg", "123.jpg", "123", "jpg", true},
		{"valid png", "456.png", "456", "png", true},
		{"valid gif", "789.gif", "789", "gif", true},
		{"uppercase extension", "100.JPG", "100", "jpg", true},
		{"mixed case extension", "200.PnG", "200", "png", true},
		{"invalid extension", "123.txt", "123", "txt", true},
		{"no extension", "123", "", "", false},
		{"empty string", "", "", "", false},
		{"non-numeric key", "abc.jpg", "", "", false},
		{"with extra text", "123-thumb.jpg", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key, ext, ok := globImage(tc.filename)
			require.Equal(t, tc.wantOk, ok)
			if ok {
				require.Equal(t, tc.wantKey, key)
				require.Equal(t, tc.wantExt, ext)
			}
		})
	}
}

func TestDetectFormat(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		buf       []byte
		wantExt   string
		wantError bool
	}{
		// JPEG magic number (FF D8 FF)
		{"JPEG header", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}, "jpg", false},
		// PNG magic number (89 50 4E 47 0D 0A 1A 0A)
		{"PNG header", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "png", false},
		// GIF magic number (47 49 46 38)
		{"GIF87a header", []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}, "gif", false},
		{"GIF89a header", []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}, "gif", false},
		// BMP magic number (42 4D)
		{"BMP header", []byte{0x42, 0x4D, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, "bmp", false},
		{"empty buffer", []byte{}, "", true},
		{"unknown format", []byte{0x00, 0x00, 0x00, 0x00}, "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ext, err := DetectFormat(tc.buf)
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantExt, ext)
			}
		})
	}
}

func TestManifestEntryMethods(t *testing.T) {
	t.Parallel()

	t.Run("Slot calculates correct slot", func(t *testing.T) {
		t.Parallel()

		entry := &ManifestEntry{
			Position: 1,
			IsLarge:  false,
		}
		require.Equal(t, config.Slot1, entry.Slot())

		entry.Position = 2
		require.Equal(t, config.Slot2, entry.Slot())

		entry.Position = 1
		entry.IsLarge = true
		require.Equal(t, config.SlotLarge1, entry.Slot())
	})

	t.Run("Comment generates description", func(t *testing.T) {
		t.Parallel()

		entry := &ManifestEntry{
			Position: 1,
			IsLarge:  false,
		}
		require.Equal(t, "Readerware image 1", entry.Comment())

		entry.Position = 3
		entry.IsLarge = true
		require.Equal(t, "Readerware large image 3", entry.Comment())
	})

	t.Run("Use marks entry as used", func(t *testing.T) {
		t.Parallel()

		entry := &ManifestEntry{
			IsUsed: false,
		}
		result := entry.Use()
		require.True(t, result.IsUsed)
		require.Same(t, entry, result) // returns same instance
	})
}

func TestRowMethods(t *testing.T) {
	t.Parallel()

	t.Run("Slot returns entry for valid slot", func(t *testing.T) {
		t.Parallel()

		entry := &ManifestEntry{}
		row := Row{}
		row[config.Slot1] = entry

		result := row.Slot(config.Slot1)
		require.Same(t, entry, result)
	})

	t.Run("Slot returns nil for invalid slot", func(t *testing.T) {
		t.Parallel()

		row := Row{}
		result := row.Slot(config.Slot(99))
		require.Nil(t, result)
	})

	t.Run("Slot returns nil for empty slot", func(t *testing.T) {
		t.Parallel()

		row := Row{}
		result := row.Slot(config.Slot1)
		require.Nil(t, result)
	})

	t.Run("Image returns entry for valid position", func(t *testing.T) {
		t.Parallel()

		entry := &ManifestEntry{}
		row := Row{}
		row[config.Slot1] = entry

		result := row.Image(1)
		require.Same(t, entry, result)
	})

	t.Run("Image returns nil for invalid position", func(t *testing.T) {
		t.Parallel()

		row := Row{}
		require.Nil(t, row.Image(0))
		require.Nil(t, row.Image(5))
	})

	t.Run("LargeImage returns entry for valid position", func(t *testing.T) {
		t.Parallel()

		entry := &ManifestEntry{}
		row := Row{}
		row[config.SlotLarge1] = entry

		result := row.LargeImage(1)
		require.Same(t, entry, result)
	})

	t.Run("LargeImage returns nil for invalid position", func(t *testing.T) {
		t.Parallel()

		row := Row{}
		require.Nil(t, row.LargeImage(0))
		require.Nil(t, row.LargeImage(5))
	})

	t.Run("Cover returns first non-nil entry", func(t *testing.T) {
		t.Parallel()

		entry1 := &ManifestEntry{}
		entry2 := &ManifestEntry{}
		row := Row{}
		row[config.Slot3] = entry1
		row[config.Slot5] = entry2

		result := row.Cover()
		require.Same(t, entry1, result) // Slot3 comes before Slot5
	})

	t.Run("Cover returns nil when no entries", func(t *testing.T) {
		t.Parallel()

		row := Row{}
		require.Nil(t, row.Cover())
	})

	t.Run("LargeCover returns first large image", func(t *testing.T) {
		t.Parallel()

		entry1 := &ManifestEntry{}
		entry2 := &ManifestEntry{}
		row := Row{}
		row[config.SlotLarge2] = entry1
		row[config.SlotLarge4] = entry2

		result := row.LargeCover()
		require.Same(t, entry1, result)
	})

	t.Run("LargeCover returns nil when no large images", func(t *testing.T) {
		t.Parallel()

		row := Row{}
		// Ensure no entries in any slot that could be inverted to find an entry
		require.Nil(t, row.LargeCover())
	})

	t.Run("First through Eighth methods", func(t *testing.T) {
		t.Parallel()

		entry1 := &ManifestEntry{}
		entry8 := &ManifestEntry{}
		row := Row{}
		row[config.Slot1] = entry1
		row[config.Slot8] = entry8

		require.Same(t, entry1, row.First())
		require.Nil(t, row.Second())
		require.Nil(t, row.Third())
		require.Nil(t, row.Fourth())
		require.Nil(t, row.Fifth())
		require.Nil(t, row.Sixth())
		require.Nil(t, row.Seventh())
		require.Same(t, entry8, row.Eighth())
	})
}

func TestIndex_All(t *testing.T) {
	t.Parallel()

	t.Run("iterates over all entries", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		imgDir := filepath.Join(tmpDir, "images")

		require.NoError(t, os.Mkdir(imgDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(imgDir, "123.jpg"), []byte("fake image 1"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(imgDir, "456.png"), []byte("fake image 2"), 0644))

		dirs := config.Directories{First: imgDir}
		index, err := BuildIndex(t.Context(), dirs)
		require.NoError(t, err)

		count := 0
		for entry := range index.All() {
			require.NotNil(t, entry)
			count++
		}
		require.Equal(t, 2, count)
	})

	t.Run("handles nil index", func(t *testing.T) {
		t.Parallel()

		var index *Index
		count := 0
		for range index.All() {
			count++
		}
		require.Equal(t, 0, count)
	})
}
