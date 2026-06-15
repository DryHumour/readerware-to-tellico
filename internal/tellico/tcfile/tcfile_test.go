package tcfile

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func fh(name string) Image {
	return Image{
		header: &zip.FileHeader{
			Name:   imagesDir + name,
			Method: zip.Store,
		},
	}
}

func TestTCFile(t *testing.T) {
	t.Run("creates valid tellico archive", func(t *testing.T) {
		var buf bytes.Buffer
		tc := New(&buf)

		// Create the main collection file
		collWriter, err := tc.Collection()
		if err != nil {
			t.Fatalf("failed to create collection: %v", err)
		}
		if _, err := io.WriteString(collWriter, "<tellico>test data</tellico>"); err != nil {
			t.Fatalf("failed to write to collection: %v", err)
		}

		// Create some images
		img1, err := tc.Image(fh("cover.jpg"))
		if err != nil {
			t.Fatalf("failed to create image 1: %v", err)
		}
		if _, err := io.WriteString(img1, "fake jpeg data"); err != nil {
			t.Fatalf("failed to write to image 1: %v", err)
		}

		img2, err := tc.Image(fh("back.png"))
		if err != nil {
			t.Fatalf("failed to create image 2: %v", err)
		}
		if _, err := io.WriteString(img2, "fake png data"); err != nil {
			t.Fatalf("failed to write to image 2: %v", err)
		}

		// Close the archive
		if err := tc.Close(); err != nil {
			t.Fatalf("failed to close tcfile: %v", err)
		}

		// Verify the zip contents
		zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if err != nil {
			t.Fatalf("failed to read zip archive: %v", err)
		}

		expectedFiles := map[string]struct {
			content string
			method  uint16
		}{
			"tellico.xml":      {content: "<tellico>test data</tellico>", method: zip.Deflate},
			"images/":          {content: "", method: zip.Store}, // Directories are usually STORE
			"images/cover.jpg": {content: "fake jpeg data", method: zip.Store},
			"images/back.png":  {content: "fake png data", method: zip.Store},
		}

		if len(zr.File) != len(expectedFiles) {
			t.Errorf("expected %d files, got %d", len(expectedFiles), len(zr.File))
			for _, f := range zr.File {
				t.Logf("found file: %s", f.Name)
			}
		}

		for _, f := range zr.File {
			expected, ok := expectedFiles[f.Name]
			if !ok {
				t.Errorf("unexpected file in archive: %s", f.Name)
				continue
			}

			// tellico.xml uses default zip method (Deflate). Our Image method explicitly uses zip.Store
			if f.Name == "tellico.xml" {
				if f.Method != zip.Deflate {
					t.Errorf("file %s has method %d, expected %d", f.Name, f.Method, expected.method)
				}
			} else {
				if f.Method != expected.method {
					t.Errorf("file %s has method %d, expected %d", f.Name, f.Method, expected.method)
				}
			}

			// Don't try to read contents of directories
			if strings.HasSuffix(f.Name, "/") {
				continue
			}

			rc, err := f.Open()
			if err != nil {
				t.Errorf("failed to open file %s: %v", f.Name, err)
				continue
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Errorf("failed to read file %s: %v", f.Name, err)
				continue
			}

			if string(content) != expected.content {
				t.Errorf("file %s content mismatch: got %q, want %q", f.Name, string(content), expected.content)
			}
		}
	})
}
