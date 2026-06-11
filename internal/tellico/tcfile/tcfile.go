package tcfile

import (
	"archive/zip"
	"fmt"
	"io"
)

const (
	// tellicoXMLFilename is the name of the XML file inside the Tellico .tc archive.
	tellicoXMLFilename = "tellico.xml"
	// imagesDir is the directory name for images inside the Tellico .tc archive.
	imagesDir = "images/"
)

// TCFile represents a Tellico zip archive being written.
type TCFile struct {
	zipWriter     *zip.Writer
	imagesCreated bool
}

// New creates a new TCFile writing to the provided io.Writer.
func New(w io.Writer) *TCFile {
	return &TCFile{
		zipWriter: zip.NewWriter(w),
	}
}

// Collection returns an io.Writer for the main tellico.xml file inside the archive.
// This should typically be called once per archive.
func (t *TCFile) Collection() (io.Writer, error) {
	w, err := t.zipWriter.Create(tellicoXMLFilename)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", tellicoXMLFilename, err)
	}
	return w, nil
}

// Image returns an io.Writer for an image file inside the archive.
// It lazily creates the images directory the first time it is called.
// Images are stored without compression (zip.Store) as they are typically already compressed.
func (t *TCFile) Image(image Image) (io.Writer, error) {
	if !t.imagesCreated {
		if _, err := t.zipWriter.Create(imagesDir); err != nil {
			return nil, fmt.Errorf("create images directory: %w", err)
		}
		t.imagesCreated = true
	}
	w, err := t.zipWriter.CreateHeader(image.header)
	if err != nil {
		return nil, fmt.Errorf("create image %s: %w", image.header.Name, err)
	}
	return w, nil
}

// Close finalizes the zip archive. It must be called when writing is complete.
func (t *TCFile) Close() error {
	if err := t.zipWriter.Close(); err != nil {
		return fmt.Errorf("finalize tc: %w", err)
	}
	return nil
}
