package normalize

import (
	"regexp"
	"slices"
)

// Markers implements a configurable marker normalization pipeline.
//
// The pipeline is designed to be executed in a specific semantic order:
// 1. Filter (Keep/Discard)
// 2. Scrub (Remove noise)
// 3. Extract (Pull out markers)
// 4. Audit (Flag potential issues)
//
// While methods are exported for flexibility, they have expectations
// about the state of the input string to ensure correct results.
type Markers struct {
	// config is a clone of the original configuration.
	config MarkersConfig
	// keepRE is a regexp matching the keep patterns.
	keepRE *regexp.Regexp
	// discardRE is a regexp matching the discard patterns.
	discardRE *regexp.Regexp
	// scrubRE is a regexp matching patterns to scrub.
	scrubRE *regexp.Regexp
	// markers is a map of markers to canonical markers.
	markers map[string]string
	// markersRE is a regexp matching the marker patterns.
	markersRE *regexp.Regexp
	// auditRE is a regexp matching the audit patterns.
	auditRE *regexp.Regexp
	// auditExplanations is a map of lowercase audit patterns to explanatory strings.
	auditExplanations map[string]string
}

// NewMarkers compiles the provided configuration into a Markers normalizer ready for use.
// Empty or nil slices/maps result in the corresponding matching logic being disabled.
func NewMarkers(config MarkersConfig) (*Markers, error) {
	var err error
	result := &Markers{
		config: config.Clone(),
	}
	result.keepRE, err = keepRE(result.config.Keep)
	if err != nil {
		return nil, err
	}
	result.discardRE, err = discardRE(result.config.Discard)
	if err != nil {
		return nil, err
	}
	result.scrubRE, err = basicRE("scrub", result.config.Scrub)
	if err != nil {
		return nil, err
	}
	result.markersRE, result.markers, err = canonicalRE("marker", result.config.Marker)
	if err != nil {
		return nil, err
	}
	result.auditRE, result.auditExplanations, err = auditRE("audit", result.config.Audit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Config returns a copy of the underlying configuration.
func (m *Markers) Config() MarkersConfig {
	return m.config.Clone()
}

// Begin creates a new Result with the given value.
//
// The input value has no preconditions; it need not even be UTF-8.
func (m *Markers) Begin(v string) Result {
	return Result{Value: v}
}

// Scrub removes the MarkerConfig.Scrub phrases from the result.
//
// The input Result.Value has no preconditions; it need not even be UTF-8.
//
// If the input is Result.Literal or Result.Discarded, it is returned
// unchanged.
// Otherwise, the result value has the MarkersConfig.Scrub phrases removed
// and has had its whitespace trimmed and squeezed as if by strutil.Squeeze.
func (m *Markers) Scrub(r Result) Result {
	if r.isImmutable() {
		return r
	}
	return scrub(r, m.scrubRE)
}

// Filter the result based on the MarkerConfig.Keep and MarkerConfig.Discard
// strings.
//
// The input Result.Value is expected to be sanitized as if by Scrub.
// In particular, it should have had its whitespace trimmed and squeezed as if
// by strutil.Squeeze.
func (m *Markers) Filter(r Result) Result {
	if r.isImmutable() {
		return r
	}
	return filter(r, m.keepRE, m.discardRE)
}

// Extract identifies and removes markers from the input Result.Value.
// Unique extracted canonical markers are appended to the Result.Markers.
//
// The input Result.Value is expected to be sanitized as if by Scrub.
// In particular, it should have had its whitespace trimmed and squeezed as if
// by strutil.Squeeze.
//
// If the input is Result.Literal or Result.Discarded, it is returned
// unchanged.
func (m *Markers) Extract(r Result) Result {
	if r.isImmutable() {
		return r
	}
	v, markers := extract(r.Value, m.markersRE, m.markers)
	r.Value = v
	if len(markers) > 0 {
		r.Markers = slices.Collect(uniq(chain(slices.Values(r.Markers), slices.Values(markers))))
	}
	return r
}

// Audit the result and add any audit reports to the result.
//
// The input Result.Value is expected to be sanitized as if by Extract.
// In particular, it should have had its whitespace trimmed and squeezed as if
// by strutil.Squeeze.
func (m *Markers) Audit(r Result) Result {
	if r.isImmutable() {
		return r
	}
	return audit(r, m.auditRE, m.auditExplanations)
}

// Process the given string through the Markers normalizer.
//
// The pipeline is designed to be executed in a specific semantic order:
// 1. Filter (Keep/Discard)
// 2. Scrub (Remove noise)
// 3. Extract (Pull out markers)
// 4. Audit (Flag potential issues)
//
// The input value has no preconditions; it need not even be UTF-8.
func (m *Markers) Process(s string) Result {
	r := m.Begin(s)
	r = m.Filter(r)
	if r.isImmutable() {
		return r
	}
	r = m.Scrub(r)
	r = m.Filter(r)
	if r.isImmutable() {
		return r
	}
	r = m.Extract(r)
	r = m.Filter(r)
	if r.isImmutable() {
		return r
	}
	r = m.Audit(r)
	return r
}
