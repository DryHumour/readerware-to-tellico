package images

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"golang.org/x/sync/errgroup"
)

var (
	ErrMultipleFormats   = errors.New("multiple image formats found")
	ErrUnsupportedFormat = errors.New("unsupported image format")

	// rowkeyRE matches filenames with a numeric ROWKEY followed by a type extension.
	rowkeyRE = regexp.MustCompile(`(?i)^(?P<rowkey>\d+)\.(?P<ext>.+)$`)
)

type Index struct {
	slot [8]map[string]*ManifestEntry
}

func BuildIndex(ctx context.Context, dirs config.Directories) (*Index, error) {
	result := &Index{}
	g, ctx := errgroup.WithContext(ctx)
	for n := range result.slot {
		slot := config.Slot(n)
		if path := dirs.Get(slot); path != "" {
			g.Go(func() error { return result.buildSlotIndex(ctx, slot, path) })
		}
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

func (i *Index) Row(rowkey string) (result Row, ok bool) {
	if i == nil {
		return Row{}, false
	}
	result = Row{}
	for n := range i.slot {
		if entry, found := i.slot[n][rowkey]; found {
			result[n] = entry
			ok = true
		}
	}
	return result, ok
}

func (i *Index) All() iter.Seq[*ManifestEntry] {
	return func(yield func(*ManifestEntry) bool) {
		if i == nil {
			return
		}
		for _, slot := range i.slot {
			for _, entry := range slot {
				if !yield(entry) {
					return
				}
			}
		}
	}
}

func (i *Index) AllUsed() iter.Seq[*ManifestEntry] {
	return func(yield func(*ManifestEntry) bool) {
		if i == nil {
			return
		}
		for _, slot := range i.slot {
			for _, entry := range slot {
				if entry.IsUsed {
					if !yield(entry) {
						return
					}
				}
			}
		}
	}
}

func (i *Index) IsEmpty() bool {
	if i == nil {
		return true
	}
	for _, slot := range i.slot {
		if len(slot) > 0 {
			return false
		}
	}
	return true
}

func readDir(ctx context.Context, dir string) ([]fs.DirEntry, error) {
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	return os.ReadDir(dir)
}

func (i *Index) buildSlotIndex(ctx context.Context, slot config.Slot, dir string) error {
	entries, err := readDir(ctx, dir)
	if err == nil && len(entries) == 1 && entries[0].IsDir() && entries[0].Name() == "Images" {
		// User passed the folder name given to Readerware>Export>Images, but it creates an Images subdir.
		dir = filepath.Join(dir, entries[0].Name())
		entries, err = readDir(ctx, dir)
	}
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}
	result := make(map[string]*ManifestEntry)
	for _, entry := range entries {
		if err := context.Cause(ctx); err != nil {
			return err
		}
		// check for ROWKEY.EXT image
		rowkey, ext, ok := globImage(entry.Name())
		if !ok {
			continue
		}
		// check for regular file
		if !entry.Type().IsRegular() {
			continue
		}
		// check for duplicate ROWKEY e.g. ROWKEY.jpg and ROWKEY.png
		path := filepath.Join(dir, entry.Name())
		if entry, ok := result[rowkey]; ok {
			return fmt.Errorf("%w: %s and %s", ErrMultipleFormats, entry.Path, path)
		}
		// get FileInfo
		if err := context.Cause(ctx); err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		// check if format is valid
		switch ext {
		case "jpg", "png", "gif":
			// trust the extension for now (Tellico will complain later if invalid)
		case "jpeg":
			ext = "jpg"
		default:
			ext, err = DetectFileFormat(path)
			if err != nil {
				return err
			}
		}
		isGIF := ext == "gif"
		if isGIF {
			ext = "png"
		}
		// create manifest entry
		result[rowkey] = &ManifestEntry{
			Path:     path,
			Position: slot.Position(),
			IsLarge:  slot.IsLarge(),
			ID:       fmt.Sprintf("%08s_%d.%s", rowkey, slot, ext),
			Format:   ext,
			IsGIF:    isGIF,
			Info:     info,
		}
	}
	i.slot[slot] = result
	return nil
}

func globImage(filename string) (rowkey, ext string, ok bool) {
	matches := rowkeyRE.FindStringSubmatch(filename)
	if len(matches) < 3 {
		return "", "", false
	}
	rowkey, ext = matches[1], strings.ToLower(matches[2])
	return rowkey, ext, true
}

func DetectFileFormat(path string) (ext string, err error) {
	r, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer r.Close()
	buf := make([]byte, 512)
	n, err := r.Read(buf) // best effort
	if err != nil {
		return "", err
	}
	ext, err = DetectFormat(buf[:n])
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, path)
	}
	return ext, nil
}

func DetectFormat(buf []byte) (ext string, err error) {
	contentType := http.DetectContentType(buf)
	mimetype, _, _ := strings.Cut(contentType, ";")
	mimetype = strings.TrimSpace(mimetype)
	switch mimetype {
	case "image/bmp":
		return "bmp", nil
	case "image/gif":
		return "gif", nil
	case "image/jpeg":
		return "jpg", nil
	case "image/png":
		return "png", nil
	case "image/svg+xml":
		return "svg", nil
	case "image/tiff":
		return "tiff", nil
	case "image/webp":
		return "webp", nil
	case "image/x-icon":
		return "ico", nil
	default:
		return "", fmt.Errorf("%s: %w", contentType, ErrUnsupportedFormat)
	}
}
