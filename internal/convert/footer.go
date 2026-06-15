package convert

import "iter"

// FooterData contains the data needed to render the footer template.
type FooterData struct {
	Images iter.Seq[ImageData]
}

// ImageData represents an image for template rendering.
type ImageData struct {
	ID     string
	Format string
}
