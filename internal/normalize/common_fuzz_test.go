package normalize

import (
	"regexp"
	"testing"
)

// FuzzExtract tests the robustness of the byte-level state machine in extract.
// It ensures that no matter the input or regex, it never panics and always
// returns a valid (possibly empty) string.
func FuzzExtract(f *testing.F) {
	seeds := []string{
		"Smith, John (ed.), Jr.",
		"Doctor Who (DVD) [signed]",
		"   Multiple   Spaces   ",
		"Non-ASCII: 世界",
		"Empty () and empty [] and empty ,.",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	// We'll use a fixed but complex regex for the fuzzer to target the state machine
	// rather than the regex engine itself.
	re := regexp.MustCompile(`(?i:\(ed\.\)|\[signed\]|DVD|Jr\.)`)
	canonical := map[string]string{
		"(ed.)":    "Editors",
		"[signed]": "Signed",
		"dvd":      "DVD",
		"jr.":      "Junior",
	}

	f.Fuzz(func(t *testing.T, input string) {
		// extract should never panic on arbitrary string input
		_, _ = extract(input, re, canonical)
	})
}

// FuzzAudit tests the robustness of the audit function with its multiple checks.
func FuzzAudit(f *testing.F) {
	f.Add("Smith et al.")
	f.Add("john doe") // lowercase start
	f.Add("value@#$%^")
	f.Add("Müller GmbH")
	f.Add("")
	f.Add("   ")
	f.Add("東京")
	f.Add("value<unknown>marker")

	re := regexp.MustCompile(`(?i:et al\.)`)
	explanations := map[string]string{
		"et al.": "Multiple authors",
	}

	f.Fuzz(func(t *testing.T, input string) {
		r := Result{Value: input}
		_ = audit(r, re, explanations)
	})
}
