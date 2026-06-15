package collection

import (
	"errors"
	"fmt"
	"maps"

	"github.com/DryHumour/readerware-to-tellico/internal/normalize"
	"github.com/DryHumour/readerware-to-tellico/internal/strutil"
)

var (
	_ CollectionInfo = (*collectionInfo)(nil)
)

type collectionInfo struct {
	// kind is the kind of collection this configuration is for.
	kind Kind
	// templateNames are the template names for this collection and policy.
	templateNames TemplateNames
	// columns holds the column configuration for this collection and policy.
	columns ColumnConfig
	// names holds the name configuration for this collection and policy.
	names normalize.NamesConfig
	// markers holds the marker configuration for this collection and policy.
	markers normalize.MarkersConfig
	// normalize holds the normalizers for this collection and policy.
	normalize Normalize
	// blocklist holds the category node blocklist for this collection and policy.
	blocklist map[string]bool
}

func (c *collectionInfo) Kind() Kind                   { return c.kind }
func (c *collectionInfo) TemplateNames() TemplateNames { return c.templateNames }
func (c *collectionInfo) Columns() ColumnConfig        { return c.columns.Clone() }
func (c *collectionInfo) Blocklist() map[string]bool   { return maps.Clone(c.blocklist) }
func (c *collectionInfo) Data() any                    { return (*collectionInfoDataView)(c) }

func (c *collectionInfo) Normalize() (result Normalize, err error) {
	var errs []error
	if c.normalize.Names == nil || c.normalize.Markers == nil {
		//fmt.Printf("realBooksConfigNames = %#v\n", c.names) // FIXME(nschelle)
		//fmt.Printf("realBooksConfigMarkers = %#v\n", c.markers) // FIXME(nschelle)
		if c.normalize.Names, err = normalize.NewNames(c.names); err != nil {
			errs = append(errs, fmt.Errorf("failed to create names normalizer: %w", err))
		}
		if c.normalize.Markers, err = normalize.NewMarkers(c.markers); err != nil {
			errs = append(errs, fmt.Errorf("failed to create markers normalizer: %w", err))
		}
	}
	return c.normalize, errors.Join(errs...)
}

type collectionInfoDataView collectionInfo

func (c *collectionInfoDataView) Columns() *collectionInfoColumnView {
	return (*collectionInfoColumnView)(c)
}

func (c *collectionInfoDataView) Blocklist(v any) (_ Nothing, err error) {
	c.blocklist, err = asBoolMap("Blocklist", v)
	return
}

func (c *collectionInfoDataView) Names() *collectionInfoNameConfigView {
	return (*collectionInfoNameConfigView)(c)
}

func (c *collectionInfoDataView) Markers() *collectionInfoMarkerConfigView {
	return (*collectionInfoMarkerConfigView)(c)
}

type collectionInfoColumnView collectionInfo

func (c *collectionInfoColumnView) Names(v any) (_ Nothing, err error) {
	c.columns.Names, err = asStringSliceMap("Columns.Names", v)
	return
}

func (c *collectionInfoColumnView) Markers(v any) (_ Nothing, err error) {
	c.columns.Markers, err = asBoolMap("Columns.Markers", v)
	return
}

func (c *collectionInfoColumnView) Categories(v any) (_ Nothing, err error) {
	c.columns.Categories, err = asBoolMap("Columns.Categories", v)
	return
}

type collectionInfoNameConfigView collectionInfo

// Keep sets the verbatim pass-through names.
func (c *collectionInfoNameConfigView) Keep(v any) (_ Nothing, err error) {
	c.names.Keep, err = asSlice("Names.Keep", v)
	return
}

// Discard sets the names to silently discard.
func (c *collectionInfoNameConfigView) Discard(v any) (_ Nothing, err error) {
	c.names.Discard, err = asSlice("Names.Discard", v)
	return
}

// Scrub sets phrases to strip before name processing.
func (c *collectionInfoNameConfigView) Scrub(v any) (_ Nothing, err error) {
	c.names.Scrub, err = asSlice("Names.Scrub", v)
	return
}

// Role sets the role-to-canonical-role mapping.
func (c *collectionInfoNameConfigView) Role(v any) (_ Nothing, err error) {
	c.names.Role, err = asMap("Names.Role", v)
	return
}

// Marker sets the marker-to-canonical-marker mapping.
func (c *collectionInfoNameConfigView) Marker(v any) (_ Nothing, err error) {
	c.names.Marker, err = asMap("Names.Marker", v)
	return
}

// Suffixes sets the name suffixes (e.g. "Jr.", "III").
func (c *collectionInfoNameConfigView) Suffixes(v any) (_ Nothing, err error) {
	c.names.Suffixes, err = asSlice("Names.Suffixes", v)
	return
}

// Honorifics sets the name honorifics (e.g. "Dr.", "Prof.").
func (c *collectionInfoNameConfigView) Honorifics(v any) (_ Nothing, err error) {
	c.names.Honorifics, err = asSlice("Names.Honorifics", v)
	return
}

// Corporate sets the corporate signifiers (e.g. "Inc.", "Corp.").
func (c *collectionInfoNameConfigView) Corporate(v any) (_ Nothing, err error) {
	c.names.Corporate, err = asSlice("Names.Corporate", v)
	return
}

// Collaboration sets the collaboration signifiers (e.g. "and", "with").
func (c *collectionInfoNameConfigView) Collaboration(v any) (_ Nothing, err error) {
	c.names.Collaboration, err = asSlice("Names.Collaboration", v)
	return
}

// Audit sets the names that trigger an audit report.
// The argument should be a map of pattern strings to explanatory strings.
func (c *collectionInfoNameConfigView) Audit(v any) (_ Nothing, err error) {
	c.names.Audit, err = asMap("Names.Audit", v)
	return
}

type collectionInfoMarkerConfigView collectionInfo

// Keep sets the verbatim pass-through names.
func (c *collectionInfoMarkerConfigView) Keep(v any) (_ Nothing, err error) {
	c.markers.Keep, err = asSlice("Markers.Keep", v)
	return
}

// Discard sets the names to silently discard.
func (c *collectionInfoMarkerConfigView) Discard(v any) (_ Nothing, err error) {
	c.markers.Discard, err = asSlice("Markers.Discard", v)
	return
}

// Scrub sets phrases to strip before marker processing.
func (c *collectionInfoMarkerConfigView) Scrub(v any) (_ Nothing, err error) {
	c.markers.Scrub, err = asSlice("Markers.Scrub", v)
	return
}

// Marker sets the marker-to-canonical-marker mapping.
func (c *collectionInfoMarkerConfigView) Marker(v any) (_ Nothing, err error) {
	c.markers.Marker, err = asMap("Markers.Marker", v)
	return
}

// Audit sets the markers that trigger an audit report.
// The argument should be a map of pattern strings to explanatory strings.
func (c *collectionInfoMarkerConfigView) Audit(v any) (_ Nothing, err error) {
	c.markers.Audit, err = asMap("Markers.Audit", v)
	return
}

// asMap converts v to a map[string]string with error wrapping.
func asMap(name string, v any) (map[string]string, error) {
	m, err := strutil.ToStringStringMap(v)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return m, nil
}

// asStringSliceMap converts v to a map[string][]string with error wrapping and validation.
// Validates that no column name appears in more than one role's column list.
func asStringSliceMap(name string, v any) (map[string][]string, error) {
	m, err := strutil.ToStringStringSliceMap(v)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	// Validate that no column appears in multiple roles
	columnToRoles := make(map[string]string)
	for role, columns := range m {
		for _, col := range columns {
			if existingRole, ok := columnToRoles[col]; ok {
				return nil, fmt.Errorf("%s: column %q appears in both %q and %q", name, col, existingRole, role)
			}
			columnToRoles[col] = role
		}
	}
	return m, nil
}

// asSlice converts v to a []string with error wrapping.
func asSlice(name string, v any) ([]string, error) {
	s, err := strutil.ToStringSlice(v)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return s, nil
}

// asBoolMap converts v to a map[string]bool with error wrapping.
func asBoolMap(name string, v any) (map[string]bool, error) {
	s, err := strutil.ToStringSlice(v)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	m := make(map[string]bool, len(s))
	for _, key := range s {
		m[key] = true
	}
	return m, nil
}
