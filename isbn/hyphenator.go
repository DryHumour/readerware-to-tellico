package isbn

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrMissingISBNRangeData      = errors.New("missing ISBN range data")
	ErrRegistrationGroupNotFound = errors.New("registration group not found")
	ErrPublisherRangeNotFound    = errors.New("invalid or unassigned publisher range")
	ErrInvalidPublisherLength    = errors.New("invalid publisher length")
)

type compiledRule struct {
	low       int
	high      int
	pubLen    int
	prefixLen int
}

type compiledGroup struct {
	group    string
	groupLen int
	rules    map[int][]compiledRule // key: prefixLen
}

// Hyphenator is an optimized, read-only representation of ISBN range data for fast hyphenation.
// Construct it once via NewHyphenator and reuse it for many ISBNs.
//
// It assumes the underlying ISBNRanges data is valid per the official ISBN range message.
// If the data is malformed, hyphenation may return an error.
//
// NOTE: This does not currently attempt to support multiple EAN prefix lengths; it keys by the
// EAN prefix string as present in the range data (e.g. "978" or "979").
// The hyphenation algorithm remains data-driven.
type Hyphenator struct {
	groupsByEAN map[string][]compiledGroup
}

// NewHyphenator compiles ISBN range data into an optimized lookup structure.
// It assumes the underlying ISBNRanges data is valid per the official ISBN range message.
// If the data is malformed, hyphenation may return an error.
func NewHyphenator(data ISBNRanges) (*Hyphenator, error) {
	h := &Hyphenator{groupsByEAN: make(map[string][]compiledGroup)}

	for _, g := range data.ISBNRangeMessage.RegistrationGroups.Group {
		parts := strings.SplitN(g.Prefix, "-", 2)
		if len(parts) != 2 {
			continue
		}
		ean, grp := parts[0], parts[1]
		if ean == "" || grp == "" {
			continue
		}

		cg := compiledGroup{
			group:    grp,
			groupLen: len(grp),
			rules:    make(map[int][]compiledRule),
		}

		for _, r := range g.Rules.Rule {
			bounds := strings.Split(r.Range, "-")
			if len(bounds) != 2 {
				continue
			}
			loS, hiS := bounds[0], bounds[1]
			if loS == "" || hiS == "" || len(loS) != len(hiS) {
				continue
			}
			pubLen, err := strconv.Atoi(r.Length)
			if err != nil || pubLen <= 0 {
				continue
			}
			lo, err := strconv.Atoi(loS)
			if err != nil {
				continue
			}
			hi, err := strconv.Atoi(hiS)
			if err != nil {
				continue
			}
			if lo > hi {
				continue
			}

			pl := len(loS)
			cg.rules[pl] = append(cg.rules[pl], compiledRule{low: lo, high: hi, pubLen: pubLen, prefixLen: pl})
		}

		for pl := range cg.rules {
			rules := cg.rules[pl]
			sort.Slice(rules, func(i, j int) bool { return rules[i].low < rules[j].low })
			cg.rules[pl] = rules
		}

		h.groupsByEAN[ean] = append(h.groupsByEAN[ean], cg)
	}

	for ean := range h.groupsByEAN {
		groups := h.groupsByEAN[ean]
		sort.Slice(groups, func(i, j int) bool { return groups[i].groupLen > groups[j].groupLen })
		h.groupsByEAN[ean] = groups
	}

	if len(h.groupsByEAN) == 0 {
		return nil, fmt.Errorf("%w", ErrMissingISBNRangeData)
	}

	return h, nil
}

// Hyphenate applies ISBN hyphenation rules to the given ISBN.
// It assumes the underlying ISBNRanges data is valid per the official ISBN range message.
// If the data is malformed, hyphenation may return an error.
func (h *Hyphenator) Hyphenate(isbn ISBN) (string, error) {
	var errs []error

	kind, candidate, err := ParseISBN(ISBNKindAny, string(isbn))
	switch {
	case err == nil:
	case errors.Is(err, ErrInvalidCheckDigit):
		// best effort hyphenation
		errs = append(errs, err)
	default:
		return "", err
	}
	var clean string
	switch kind {
	case ISBNKind10:
		clean = string(ISBN10Prefix + candidate)
	case ISBNKind13:
		clean = string(candidate)
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidKind, kind) // (cannot occur)
	}

	if h == nil {
		return "", ErrMissingISBNRangeData
	}

	var (
		matchedEAN   string
		matchedGroup *compiledGroup
	)

	for ean, groups := range h.groupsByEAN {
		if !strings.HasPrefix(clean, ean) {
			continue
		}
		rest := clean[len(ean):]
		for idx := range groups {
			g := &groups[idx]
			if strings.HasPrefix(rest, g.group) {
				matchedEAN = ean
				matchedGroup = g
				break
			}
		}
		if matchedGroup != nil {
			break
		}
	}

	if matchedGroup == nil {
		return "", fmt.Errorf("%w: %s", ErrRegistrationGroupNotFound, clean)
	}

	rest := clean[len(matchedEAN):]
	publisherPart := rest[len(matchedGroup.group):]

	// Try prefix lengths in increasing order (shorter prefixes are more general). This matches
	// the usual interpretation of the range message rules.
	prefixLens := make([]int, 0, len(matchedGroup.rules))
	for pl := range matchedGroup.rules {
		prefixLens = append(prefixLens, pl)
	}
	sort.Ints(prefixLens)

	var rule *compiledRule
	for _, pl := range prefixLens {
		if len(publisherPart) <= pl {
			continue
		}
		x, err := strconv.Atoi(publisherPart[:pl])
		if err != nil {
			continue
		}
		rules := matchedGroup.rules[pl]
		idx := sort.Search(len(rules), func(i int) bool { return rules[i].low > x })
		if idx == 0 {
			continue
		}
		cand := rules[idx-1]
		if x >= cand.low && x <= cand.high {
			rule = &cand
			break
		}
	}

	if rule == nil {
		return "", fmt.Errorf("%w: %s", ErrPublisherRangeNotFound, clean)
	}

	pubLen := rule.pubLen
	if pubLen < 1 || pubLen >= len(publisherPart) {
		return "", fmt.Errorf("%w: %s: length %d", ErrInvalidPublisherLength, clean, pubLen)
	}

	publisher := publisherPart[:pubLen]
	publication := publisherPart[pubLen : len(publisherPart)-1]
	checkDigit := clean[len(clean)-1:]

	if kind == ISBNKind10 {
		return fmt.Sprintf("%s-%s-%s-%s", matchedGroup.group, publisher, publication, checkDigit), errors.Join(errs...)
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", matchedEAN, matchedGroup.group, publisher, publication, checkDigit), errors.Join(errs...)
}
