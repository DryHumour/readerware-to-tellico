package tcfile

import (
	"archive/zip"
	"fmt"
	"io/fs"
)

// Image describes a single image entry to be stored in the TC file archive.
type Image struct {
	header *zip.FileHeader
}

// NewImage creates an Image for the given image id, using fi for the archive
// entry metadata and comment for the zip entry comment.
func NewImage(id string, fi fs.FileInfo, comment string) (Image, error) {
	header, err := zip.FileInfoHeader(fi)
	if err != nil {
		return Image{}, fmt.Errorf("create image header: %w", err)
	}
	header.Name = imagesDir + id
	header.Method = zip.Store
	header.Comment = comment
	return Image{header: header}, nil
}
