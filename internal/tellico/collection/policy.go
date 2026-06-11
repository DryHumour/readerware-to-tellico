package collection

import (
	"context"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

// Policy defines kind-specific behavior used by the generic conversion pipeline.
type Policy interface {
	// Info returns information about this collection / policy.
	Info() CollectionInfo
	// ConfigureHeaders validates that the CSV export contains all columns required by
	// this policy and that image-related constraints are satisfied, then stores the headers.
	ConfigureHeaders(headers []string, haveImages bool) error
	// NewEntry constructs a single entry from the cleaned row data.
	NewEntry(clean map[string]string, img images.Row) (Entry, error)
}

// CollectionInfo defines the configuration for a collection / policy.
// CollectionInfo is the interface that config templates receive.
type CollectionInfo interface {
	// Kind returns the collection kind this policy handles (Books, Music, or Video).
	Kind() Kind
	// TemplateNames returns the template names used by this policy.
	TemplateNames() TemplateNames
	// Columns returns the column configuration for this collection.
	Columns() ColumnConfig
	// Normalize returns the normalizers for this policy.
	Normalize() (Normalize, error)
	// Blocklist returns the blocklist for this policy.
	Blocklist() map[string]bool
	// Data returns the data transfer object for configuring this collection / policy.
	Data() any
}

// TemplateNames holds the template names for a kind-specific policy.
// Each name corresponds to a named template executed by the conversion pipeline.
type TemplateNames struct {
	// Config is the template executed once before CSV processing to populate a EntryConfig.
	// The template output is discarded; only its side effects matter.
	Config string
	// Header is the template used to render the Tellico XML collection header.
	Header string
	// Entry is the template used to render a single Tellico XML entry.
	Entry string
	// Footer is the template used to render the Tellico XML collection footer.
	Footer string
}

// New returns a kind-specific policy for conversion.
func New(ctx context.Context, cfg *config.Config, kind Kind) (Policy, error) {
	switch kind {
	case KindBooks:
		return NewBooksPolicy(ctx, cfg), nil
	case KindMusic:
		return NewMusicPolicy(ctx, cfg), nil
	case KindVideo:
		return NewVideoPolicy(ctx, cfg), nil
	default:
		return nil, ErrUnknownKind(kind)
	}
}
