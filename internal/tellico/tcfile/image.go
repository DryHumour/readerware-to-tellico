package tcfile

import (
	"archive/zip"
	"fmt"
	"io/fs"
)

type Image struct {
	header *zip.FileHeader
}

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
