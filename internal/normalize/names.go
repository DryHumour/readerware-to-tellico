package normalize

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/DryHumour/readerware-to-tellico/internal/strutil"
)

// Names implements a multi-stage configurable name normalization pipeline.
//
// The pipeline is designed to be executed in a specific semantic order:
// 1. Filter (Keep/Discard)
// 2. Scrub (Remove noise/canonicalize suffixes)
// 3. Extract (Pull out roles and markers)
// 4. NaturalOrder (Flip "Last, First" to "First Last")
// 5. Audit (Flag potential issues)
//
// While methods are exported for flexibility, they have expectations
// about the state of the input string to ensure correct results.
type Names struct {
	// config is a clone of the original configuration.
	config NamesConfig
	// keepRE is a regexp matching the keep patterns.
	keepRE *regexp.Regexp
	// discardRE is a regexp matching the discard patterns.
	discardRE *regexp.Regexp
	// scrubRE is a regexp matching patterns to scrub.
	scrubRE *regexp.Regexp
	// role is a map of role patterns to canonical roles.
	roles map[string]string
	// roleRE is a regexp matching the role patterns.
	rolesRE *regexp.Regexp
	// marker is a map of marker text to canonical markers.
	markers map[string]string
	// markerRE is a regexp matching the marker patterns.
	markersRE *regexp.Regexp
	// suffixRE is a regexp matching the suffix patterns at end of string.
	suffixRE *regexp.Regexp
	// suffixes is a map of suffix phrases to canonical suffixes.
	suffixes map[string]string
	// ambiguousSuffixesRE is a regexp matching suffix patterns that could be part of a name.
	ambiguousSuffixesRE *regexp.Regexp
	// honorificRE is a regexp matching honorific patterns.
	honorificRE *regexp.Regexp
	// ambiguousHonorificsRE is a regexp matching honorifcs and suffix patterns within names.
	ambiguousHonorificsRE *regexp.Regexp
	// corporateRE is a regexp matching corporate signifiers.
	corporateRE *regexp.Regexp
	// punctRE is a regexp matching known trailing suffixes or corporate signifiers.
	punctRE *regexp.Regexp
	// collaborationRE is a regexp matching collaboration signifiers.
	collaborationRE *regexp.Regexp
	// auditRE is a regexp matching the audit patterns.
	auditRE *regexp.Regexp
	// auditExplanations is a map of lowercase audit patterns to explanatory strings.
	auditExplanations map[string]string
}

// NewNames compiles the provided configuration into a Names normalizer ready for use.
// Empty or nil slices/maps result in the corresponding matching logic being disabled.
func NewNames(cfg NamesConfig) (*Names, error) {
	var err error
	result := &Names{
		config: cfg.Clone(),
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
	result.rolesRE, result.roles, err = canonicalRE("role", result.config.Role)
	if err != nil {
		return nil, err
	}
	result.markersRE, result.markers, err = canonicalRE("marker", result.config.Marker)
	if err != nil {
		return nil, err
	}
	result.corporateRE, err = basicRE("corporate", result.config.Corporate)
	if err != nil {
		return nil, err
	}
	result.collaborationRE, err = basicRE("collaboration", result.config.Collaboration)
	if err != nil {
		return nil, err
	}
	result.auditRE, result.auditExplanations, err = auditRE("audit", result.config.Audit)
	if err != nil {
		return nil, err
	}
	return result.honorificAndSuffixInit()
}

// Config returns a copy of the underlying configuration.
func (n *Names) Config() NamesConfig {
	return n.config.Clone()
}

// Begin creates a new Result with the given value.
//
// The input value has no preconditions; it need not even be UTF-8.
func (n *Names) Begin(v string) Result {
	return Result{Value: v}
}

// Scrub removes the NamesConfig.Scrub phrases from the result.
//
// The input Result.Value has no preconditions; it need not even be UTF-8.
//
// If the input is Result.Literal or Result.Discarded, it is returned
// unchanged.
// Otherwise, the result value has the NamesConfig.Scrub phrases removed
// and has had its whitespace trimmed and squeezed as if by strutil.Squeeze.
func (n *Names) Scrub(r Result) Result {
	if r.isImmutable() {
		return r
	}
	r = scrub(r, n.scrubRE)
	if n.suffixRE != nil {
		if loc := n.suffixRE.FindStringIndex(r.Value); loc != nil {
			r.Value = r.Value[:loc[0]] + n.canonicalSuffix(r.Value[loc[0]:])
		}
	}
	return r
}

// Filter the result based on the NamesConfig.Keep and NamesConfig.Discard
// strings.
//
// The input Result.Value is expected to be sanitized as if by Scrub.
// In particular, it should have had its whitespace trimmed and squeezed as if
// by strutil.Squeeze.
func (n *Names) Filter(r Result) Result {
	if r.isImmutable() {
		return r
	}
	return filter(r, n.keepRE, n.discardRE)
}

// Extract identifies and removes roles and markers from the input
// Result.Value.
// Unique extracted canonical roles are appended to the Result.Roles.
// Unique extracted canonical markers are appended to the Result.Markers.
//
// The input Result.Value is expected to be sanitized as if by Scrub.
// In particular, it should have had its whitespace trimmed and squeezed as if
// by strutil.Squeeze.
//
// If the input is Result.Literal or Result.Discarded, it is returned
// unchanged.
func (n *Names) Extract(r Result) Result {
	if r.isImmutable() {
		return r
	}
	v, roles := extract(r.Value, n.rolesRE, n.roles)
	r.Value = v
	if len(roles) > 0 {
		r.Roles = slices.Collect(uniq(chain(slices.Values(r.Roles), slices.Values(roles))))
	}
	v, markers := extract(r.Value, n.markersRE, n.markers)
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
func (n *Names) Audit(r Result) Result {
	if r.isImmutable() {
		return r
	}

	if n.ambiguousHonorificsRE != nil {
		if locs := n.ambiguousHonorificsRE.FindStringIndex(r.Value); locs != nil && locs[0] != 0 {
			r.RequiresAudit = true
			r.AuditReasons = append(r.AuditReasons, AuditReasonWithMatch("[ambiguous] Name contains ambiguous honorifics", r.Value, locs))
		}
	}

	if n.ambiguousSuffixesRE != nil {
		if locs := n.ambiguousSuffixesRE.FindStringIndex(r.Value); locs != nil && locs[1] < len(r.Value) {
			r.RequiresAudit = true
			r.AuditReasons = append(r.AuditReasons, AuditReasonWithMatch("[ambiguous] Name contains ambiguous initials", r.Value, locs))
		}
	}

	if n.collaborationRE != nil {
		if locs := n.collaborationRE.FindStringIndex(r.Value); locs != nil {
			r.RequiresAudit = true
			r.AuditReasons = append(r.AuditReasons, AuditReasonWithMatch("[collaboration] Name contains collaboration signifier", r.Value, locs))
		}
	}

	if strings.ContainsAny(r.Value, "()[]{}") {
		r.RequiresAudit = true
		r.AuditReasons = append(r.AuditReasons, AuditReason("[parentheses] Name contains parentheses", r.Value))
	}

	if lastRune, _ := utf8.DecodeLastRuneInString(r.Value); unicode.IsPunct(lastRune) {
		if n.punctRE == nil || !n.punctRE.MatchString(r.Value) {
			r.RequiresAudit = true
			r.AuditReasons = append(r.AuditReasons, AuditReason("[punctuation] Name ends with punctuation", r.Value))
		}
	}

	return audit(r, n.auditRE, n.auditExplanations)
}

// Process the given string through the Names normalizer.
//
// The pipeline is designed to be executed in a specific semantic order:
// 1. Filter (Keep/Discard)
// 2. Scrub (Remove noise/canonicalize suffixes)
// 3. Filter (Keep/Discard)
// 4. Extract (Pull out roles and markers)
// 5. Filter (Keep/Discard)
// 6. NaturalOrder (Flip "Last, First" to "First Last")
// 7. Filter (Keep/Discard)
// 8. Audit (Flag potential issues)
//
// The input value has no preconditions; it need not even be UTF-8.
func (n *Names) Process(s string) Result {
	r := n.Begin(s)
	r = n.Filter(r)
	if r.isImmutable() {
		return r
	}
	r = n.Scrub(r)
	r = n.Filter(r)
	if r.isImmutable() {
		return r
	}
	r = n.Extract(r)
	r = n.Filter(r)
	if r.isImmutable() {
		return r
	}
	r = n.NaturalOrder(r)
	r = n.Filter(r)
	if r.isImmutable() {
		return r
	}
	r = n.Audit(r)
	return r
}

// NaturalOrder converts a name from "Last, First" order to "First Last" order.
//
// It handles several common name structures:
//   - "Surnames, Givens" ⟶ "Givens Surnames"
//   - "Surnames, Givens, Suffix" ⟶ "Givens Surnames Suffix" (requires configured Suffixes)
//   - "Surnames, Givens, Honorific" ⟶ "Honorific Givens Surnames" (requires configured Honorifics)
//
// Surnames, given names, and honorifics can consist of multiple words.
//
// If the input Result.Value contains corporate signifiers from
// NamesConfig.Corporate, it will be returned as-is and flagged for audit.
//
// The input Result.Value is expected to be sanitized as if by Extract.
// In particular, it should have had its whitespace trimmed and squeezed as if
// by strutil.Squeeze.
//
// If the input is Result.Literal or Result.Discarded, it is returned
// unchanged.
func (n *Names) NaturalOrder(r Result) Result {
	if r.isImmutable() {
		return r
	}

	// 1. The Fast Path (Zero Commas)

	if !strings.Contains(r.Value, ",") {
		return auditEmptyName(r)
	}

	// 2. The Corporate Short-Circuit

	// Check the full original string for corporate signifiers.
	// If a match is found (e.g., "University of...", "Dept. of..."), do not flip, and return r.

	if n.corporateRE != nil {
		if locs := n.corporateRE.FindStringIndex(r.Value); locs != nil {
			r.RequiresAudit = true
			r.AuditReasons = append(r.AuditReasons, AuditReasonWithMatch("[corporate] possible corporate name", r.Value, locs))
			return r
		}
	}

	// 3. Split & Sanitize

	// Split r.Value by commas.  Trim whitespace from all resulting segments.

	orig := r.Value
	parts := strings.Split(r.Value, ",")

	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// Empty Check: If any segment is empty (e.g., "Smith,,John" or "Smith, John,"), discard the empty segment, and continue processing the valid ones.

	if slices.Contains(parts, "") {
		r.RequiresAudit = true
		r.AuditReasons = append(r.AuditReasons, AuditReason("[null] Empty comma segment in name", orig))
		parts = slices.DeleteFunc(parts, func(s string) bool {
			return s == ""
		})
	}

	// 4. Process based on number of parts

	switch len(parts) {
	case 0:
		// no non-empty parts, return ""
		r.Value = ""
		return auditEmptyName(r)
	case 1:
		// only one non-empty part, return it
		r.Value = parts[0]
		return r
	case 2:
		// two non-empty parts: Surnames, Givens
		var locs []int
		if n.suffixRE != nil {
			locs = n.suffixRE.FindStringIndex(parts[1])
		}
		if len(locs) > 0 {
			// joined suffix found: Surnames, Givens Suffix ⟶ Givens Surnames Suffix
			givens := strings.TrimSpace(parts[1][:locs[0]])
			suffix := n.canonicalSuffix(parts[1][locs[0]:])
			r.Value = strutil.JoinParts(givens, parts[0], suffix)
			return r
		}
		// no suffixes match, simple flip: Surnames, Givens ⟶ Givens Surnames
		r.Value = strutil.JoinParts(parts[1], parts[0])
		return r
	case 3:
		// three non-empty parts: Surnames, Givens, [Honorific|Suffix|unknown]
		var sfx []int
		if n.suffixRE != nil {
			sfx = n.suffixRE.FindStringIndex(parts[2])
		}
		if len(sfx) > 0 && sfx[0] == 0 {
			// suffix found: Surnames, Givens, Suffix ⟶ Givens Surnames Suffix
			r.Value = strutil.JoinParts(parts[1], parts[0], n.canonicalSuffix(parts[2]))
			return r
		}
		var hon []int
		if n.honorificRE != nil {
			hon = n.honorificRE.FindStringIndex(parts[2])
		}
		if len(hon) > 0 {
			// honorific found: Surnames, Givens, Honorific ⟶ Honorific Givens Surnames
			r.Value = strutil.JoinParts(parts[2], parts[1], parts[0])
			return r
		}
		// third part is neither suffix nor honorific: ambiguous, rejoin with commas (best-effort)
		r.Value = strings.Join(parts, ", ")
		r.RequiresAudit = true
		r.AuditReasons = append(r.AuditReasons, AuditReason("[ambiguous] Third part is neither suffix nor honorific", r.Value))
		return r
	default:
		// too many non-empty parts, rejoin with commas (best-effort)
		r.Value = strings.Join(parts, ", ")
		r.RequiresAudit = true
		r.AuditReasons = append(r.AuditReasons, AuditReason("[ambiguous] Too many comma segments", r.Value))
		return r
	}
}

// honorificAndSuffixInit initializes the honorific and suffix related regular expressions.
func (n *Names) honorificAndSuffixInit() (*Names, error) {
	var err error

	// Build suffix patterns; canonical suffix map; ambiguous patterns.

	estimate := 2*len(n.config.Suffixes) + 10
	n.suffixes = make(map[string]string, estimate) // canonical suffixes
	sfxPats := make(map[string]string, estimate)   // basic suffix patterns
	ambPats := make(map[string]string, estimate)   // suffix patterns with optional spaces

	for _, v := range n.config.Suffixes {
		v = strutil.Squeeze(v)
		if v == "" {
			continue
		}
		k := strings.ToLower(v)
		n.suffixes[k] = v                // canonical suffix
		sfxPats[v] = regexp.QuoteMeta(v) // basic pattern
		ambPats[v] = spacePattern(v)     // pattern with optional spaces
		if stem, ok := strings.CutSuffix(k, "."); ok {
			// add pattern for suffix without trailing period
			n.suffixes[stem] = v
			sfxPats[stem] = regexp.QuoteMeta(stem)
		}
		if spc, pat, ok := abbrevPattern(k); ok {
			// add pattern for suffix with one or more spaces
			n.suffixes[spc] = v
			sfxPats[spc] = pat
		}
	}

	// Compile suffixes regular expression (case-insensitive, match at end only).

	n.suffixRE, err = compileRE("i", ``, boundaryPatternsByLen(`\b`, maps.All(sfxPats), ``), `$`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile suffix patterns: %w", err)
	}

	n.ambiguousSuffixesRE, err = compileRE("i", ``, boundaryPatternsByLen(`\b`, maps.All(ambPats), `\b`), ``)
	if err != nil {
		return nil, fmt.Errorf("failed to compile suffix ambiguity patterns: %w", err)
	}

	// Build honorific patterns; add to ambiguous patterns.

	honPats := make(map[string]string, len(n.config.Honorifics))
	for _, v := range n.config.Honorifics {
		v = strutil.Squeeze(v)
		if v == "" {
			continue
		}
		honPats[v] = regexp.QuoteMeta(v)
	}

	// Compile honorifics regular expression (case-insensitive, match at start or end or word boundaries).

	n.honorificRE, err = compileRE("i", ``, boundaryPatternsByLen(`\b`, maps.All(honPats), `\b`), ``)
	if err != nil {
		return nil, fmt.Errorf("failed to compile honorific patterns: %w", err)
	}

	n.ambiguousHonorificsRE, err = compileRE("i", ``, boundaryPatternsByLen(`\b`, maps.All(honPats), `\b`), ``)
	if err != nil {
		return nil, fmt.Errorf("failed to compile honorific ambiguity patterns: %w", err)
	}

	// Build punctuation patterns for suffixes or corporate signifiers ending in punctuation.

	punctPat := make(map[string]string, len(n.config.Suffixes)+len(n.config.Corporate))

	for _, v := range n.config.Suffixes {
		v = strutil.Squeeze(v)
		if v == "" {
			continue
		}
		if lastRune, _ := utf8.DecodeLastRuneInString(v); unicode.IsPunct(lastRune) {
			punctPat[v] = sfxPats[v]
		}
	}

	for _, v := range n.config.Corporate {
		v = strutil.Squeeze(v)
		if v == "" {
			continue
		}
		if lastRune, _ := utf8.DecodeLastRuneInString(v); unicode.IsPunct(lastRune) {
			punctPat[v] = regexp.QuoteMeta(v)
		}
	}

	// Compile punctuation regular expression (case-insensitive, match at end only).

	n.punctRE, err = compileRE("i", ``, boundaryPatternsByLen(`\b`, maps.All(punctPat), ``), `$`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile punctuation patterns: %w", err)
	}

	return n, nil
}

// canonicalSuffix returns the canonical form of a suffix string.
func (n *Names) canonicalSuffix(s string) string {
	if c, ok := n.suffixes[strings.ToLower(strutil.Squeeze(s))]; ok {
		return c
	}
	return s
}

func auditEmptyName(r Result) Result {
	if r.Value == "" && !slices.Contains(r.Markers, AnonymousMarker) && !slices.Contains(r.Markers, UnknownMarker) {
		r.RequiresAudit = true
		r.AuditReasons = append(r.AuditReasons, "[empty] Empty name")
	}
	return r
}
