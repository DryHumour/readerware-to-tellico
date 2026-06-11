package normalize

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	realBooksConfigNames   = NamesConfig{Keep: []string{"Soundtrack", "Various", "Various Artists", "Various Authors"}, Discard: []string{"-", "be the first to review this item", "be the first to review this item.", "more by this author", "more by this author.", "n/a", "na", "none", "sorry this product is currently out of stock", "sorry this product is currently out of stock."}, Scrub: []string{"be the first to audit this item", "see more titles by", "see other works by"}, Role: map[string]string{"(contrib)": "Contributors", "(contrib.)": "Contributors", "(contributor)": "Contributors", "(cover)": "Illustrators", "(ed)": "Editors", "(ed.)": "Editors", "(editor)": "Editors", "(foreword by)": "Foreworders", "(forward)": "Foreworders", "(fwd)": "Foreworders", "(fwd.)": "Foreworders", "(il)": "Illustrators", "(il.)": "Illustrators", "(illus.)": "Illustrators", "(illustrator)": "Illustrators", "(intr)": "Introducers", "(intr.)": "Introducers", "(intro)": "Introducers", "(intro.)": "Introducers", "(introd)": "Introducers", "(introd.)": "Introducers", "(introduction by)": "Introducers", "(introduction)": "Introducers", "(pref)": "Forwarders", "(pref.)": "Forwarders", "(rev)": "Revisers", "(rev.)": "Revisers", "(revised by)": "Revisers", "(tr.)": "Translators", "(trans.)": "Translators", "(translator)": "Translators", "[contrib.]": "Contributors", "[contrib]": "Contributors", "[contributor]": "Contributors", "[cover]": "Illustrators", "[ed.]": "Editors", "[ed]": "Editors", "[editor]": "Editors", "[foreword by]": "Foreworders", "[forward]": "Foreworders", "[fwd.]": "Foreworders", "[fwd]": "Foreworders", "[il.]": "Illustrators", "[il]": "Illustrators", "[illus.]": "Illustrators", "[illustrator]": "Illustrators", "[intr.]": "Introducers", "[intr]": "Introducers", "[intro.]": "Introducers", "[intro]": "Introducers", "[introd.]": "Introducers", "[introd]": "Introducers", "[introduction by]": "Introducers", "[introduction]": "Introducers", "[pref.]": "Forwarders", "[pref]": "Forwarders", "[rev.]": "Revisers", "[rev]": "Revisers", "[revised by]": "Revisers", "[tr.]": "Translators", "[trans.]": "Translators", "[translator]": "Translators", "see more titles edited by": "Editors", "see more titles illustrated by": "Illustrators", "see more titles translated by": "Translators", "see more titles with a foreword by": "Foreworders"}, Marker: map[string]string{"(anon)": "<anonymous>", "(anon.)": "<anonymous>", "(anonym)": "<anonymous>", "(anonym.)": "<anonymous>", "(anonymous)": "<anonymous>", "(signed copy)": "<signed>", "(signed)": "<signed>", "(unk)": "<unknown>", "(unk.)": "<unknown>", "(unknown)": "<unknown>", "Anonymous": "<anonymous>", "Unknown": "<unknown>", "[anon.]": "<anonymous>", "[anon]": "<anonymous>", "[anonym.]": "<anonymous>", "[anonym]": "<anonymous>", "[anonymous]": "<anonymous>", "[signed copy]": "<signed>", "[signed]": "<signed>", "[unk.]": "<unknown>", "[unk]": "<unknown>", "[unknown]": "<unknown>"}, Suffixes: []string{"II", "III", "IV", "Jr.", "Sr.", "Esq.", "B.Sc.", "B.V.Sc.", "C.B.E.", "C.M.G.", "C.O.M.", "C.P.A.", "C.V.M.", "D.Min.", "D.Phil.", "D.V.M.", "Ed.D.", "F.B.A.", "F.R.C.P.", "F.R.G.S.", "F.R.S.", "F.R.S.C.", "F.R.S.L.", "G.B.E.", "G.C.B.", "G.C.M.G.", "K.B.E.", "K.C.B.", "K.C.M.G.", "LL.B.", "LL.D.", "LL.M.", "M.B.E.", "M.L.A.", "M.N.A.", "M.P.P.", "M.Sc.", "M.V.O.", "O.B.E.", "O.F.M.", "O.M.M.", "O.O.N.", "O.Ont.", "P.Eng.", "Ph.D.", "S.O.M.", "Th.D."}, Honorifics: []string{"The", "Dr", "Dr.", "Lady", "Lord", "Miss", "Mlle", "Mlle.", "Mme", "Mme.", "Mr", "Mr.", "Mrs", "Mrs.", "Ms", "Ms.", "Prof", "Prof.", "Professor", "Reverend", "Saint", "Sir", "St", "St.", "The Reverend", "The Rev", "Rev.", "Rev", "The Very Reverend", "The Very Rev", "The Very Rev", "Very Rev.", "Very Rev", "The Right Reverend", "The Rt. Rev.", "The Rt. Rev", "The Rt Rev", "Rt. Rev.", "Rt. Rev", "The Most Reverend", "The Most Rev", "The Most Rev", "Most Rev.", "Most Rev", "The Right Honourable", "The Rt. Hon.", "The Rt Hon", "Rt. Hon", "The Honourable", "The Hon.", "The Hon", "Hon.", "His Beatitude", "His Eminence", "Her Eminence", "His Excellency", "Her Excellency", "His Holiness"}, Corporate: []string{"Association", "Assoc.", "Assoc", "Assn.", "Committee", "Comm.", "Cttee.", "Cmte.", "Department", "Dept.", "Division", "Div.", "Federation", "Fed.", "Institute", "Inst.", "Organization", "Organisation", "Org.", "Society", "Soc.", "Press", "University", "Univ.", "U. of", "Company", "Co.", "Co", "Corporation", "Corp.", "Corp", "Incorporated", "Inc.", "Inc", "Limited", "Ltd.", "Ltd", "Band", "Choir", "Chorus", "Ensemble", "Octet", "Orchestra", "Philharmonic", "Players", "Quartet", "Quintet", "Septet", "Sextet", "Symphony", "Trio"}, Collaboration: []string{"&", "/", "+", "featuring", "feat.", "ft.", " with "}, Audit: map[string]string{"B.A.": "[audit] Bachelor of Arts OR initials - verify if suffix or given names", "B.S.": "[audit] Bachelor of Science OR initials - verify if suffix or given names", "C.A.": "[audit] Chartered Accountant OR initials - verify if suffix or given names", "C.B.": "[audit] Companion of the Bath OR initials - verify if suffix or given names", "C.C.": "[audit] Companion of the Order of Canada OR initials - verify if suffix or given names", "C.H.": "[audit] Companion of Honour OR initials - verify if suffix or given names", "C.M.": "[audit] Member of the Order of Canada OR initials - verify if suffix or given names", "C.Q.": "[audit] Knight of the National Order of Quebec OR initials - verify if suffix or given names", "D.D.": "[audit] Doctor of Divinity OR initials - verify if suffix or given names", "J.D.": "[audit] Juris Doctor OR initials - verify if suffix or given names", "J.P.": "[audit] Justice of the Peace OR initials - verify if suffix or given names", "K.C.": "[audit] King's Counsel OR initials - verify if suffix or given names", "K.G.": "[audit] Knight of the Garter OR initials - verify if suffix or given names", "K.T.": "[audit] Knight of the Thistle OR initials - verify if suffix or given names", "M.A.": "[audit] Master of Arts OR initials - verify if suffix or given names", "M.C.": "[audit] Military Cross OR initials - verify if suffix or given names", "M.D.": "[audit] Doctor of Medicine OR initials - verify if suffix or given names", "M.P.": "[audit] Member of Parliament OR initials - verify if suffix or given names", "M.S.": "[audit] Master of Science OR initials - verify if suffix or given names", "O.C.": "[audit] Officer of the Order of Canada OR initials - verify if suffix or given names", "O.M.": "[audit] Order of Merit OR initials - verify if suffix or given names", "O.P.": "[audit] Order of Preachers (Dominican) OR initials - verify if suffix or given names", "O.Q.": "[audit] Officer of the National Order of Quebec OR initials - verify if suffix or given names", "P.C.": "[audit] Privy Counsellor OR initials - verify if suffix or given names", "Q.C.": "[audit] Queen's Counsel OR initials - verify if suffix or given names", "R.A.": "[audit] Royal Academician OR initials - verify if suffix or given names", "R.N.": "[audit] Registered Nurse OR initials - verify if suffix or given names", "S.J.": "[audit] Society of Jesus (Jesuit) OR initials - verify if suffix or given names", "author": "[audit] Role marker trapped in name string - should be extracted to Roles", "authors": "[audit] Role marker trapped in name string - should be extracted to Roles", "edited by": "[audit] Role marker trapped in name string - should be extracted to Editors", "editor": "[audit] Role marker trapped in name string - should be extracted to Editors", "et al": "[audit] Multiple authors omitted ('and others') - verify if full author list is needed", "et al.": "[audit] Multiple authors omitted ('and others') - verify if full author list is needed", "illustrated by": "[audit] Role marker trapped in name string - should be extracted to Illustrators", "illustrator": "[audit] Role marker trapped in name string - should be extracted to Illustrators", "n/a": "[audit] Database placeholder detected - name data is missing", "product": "[audit] Database artifact detected - likely corrupt data entry", "sorry": "[audit] Database error string detected - verify original data source", "translated by": "[audit] Role marker trapped in name string - should be extracted to Translators", "translator": "[audit] Role marker trapped in name string - should be extracted to Translators"}}
	realBooksConfigMarkers = MarkersConfig{Keep: []string{}, Discard: []string{"-", "n/a", "na", "none", "sorry this product is currently out of stock", "sorry this product is currently out of stock."}, Scrub: []string{}, Marker: map[string]string{"(o. p.)": "<out_of_print>", "(o.p.)": "<out_of_print>", "(out of print)": "<out_of_print>", "(signed copy)": "<signed>", "(signed)": "<signed>", "[o. p.]": "<out_of_print>", "[o.p.]": "<out_of_print>", "[out of print]": "<out_of_print>", "[signed copy]": "<signed>", "[signed]": "<signed>"}, Audit: map[string]string{"et al": "[audit] Latin abbreviation meaning 'and others' - verify authorship", "et al.": "[audit] Latin abbreviation meaning 'and others' - verify authorship", "feat.": "[audit] Abbreviation for 'featuring' - verify collaboration details", "ft.": "[audit] Abbreviation for 'featuring' - verify collaboration details", "n/a": "[audit] Not applicable placeholder - verify data source"}}
)

func TestBugFixes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Result
	}{
		{
			input: "Anonymous",
			want: Result{
				Value:         "",
				Markers:       []string{AnonymousMarker},
				RequiresAudit: true,
				AuditReasons: []string{
					"[anonymous] Marker for " + AnonymousMarker + " present: \"\"",
				},
			},
		},
		{
			input: "Saint-Exupery, Antoine de",
			want: Result{
				Value:         "Antoine de Saint-Exupery",
				RequiresAudit: true,
				AuditReasons: []string{
					"[ambiguous] Name contains ambiguous honorifics: \"Saint\": \"Antoine de Saint-Exupery\"",
				},
			},
		},
		{
			input: "bner, CharleIII", // (sic)
			want: Result{
				Value: "CharleIII bner",
			},
		},
		{
			input: "Nain, Ravi",
			want: Result{
				Value: "Ravi Nain",
			},
		},
		{
			input: "Oppenheim, Alan V.",
			want: Result{
				Value: "Alan V. Oppenheim",
			},
		},
		{
			input: "Seuss, Dr",
			want: Result{
				Value: "Dr Seuss",
			},
		},
		{
			input: " Rieu, Emil V.",
			want: Result{
				Value: "Emil V. Rieu",
			},
		},
		{
			input: "Brettler, Marc Zvi",
			want: Result{
				Value: "Marc Zvi Brettler",
			},
		},
		{
			input: "Walker, William O. Jr.",
			want: Result{
				Value: "William O. Walker Jr.",
			},
		},
		{
			input: "Guile Developers, The",
			want: Result{
				Value: "The Guile Developers",
			},
		},
		{
			input: "Vaughn, Gary V.",
			want: Result{
				Value: "Gary V. Vaughn",
			},
		},
		{
			input: "Deshpande, Madhav",
			want: Result{
				Value: "Madhav Deshpande",
			},
		},
		{
			input: "Verner, Miroslav",
			want: Result{
				Value: "Miroslav Verner",
			},
		},
		{
			input: "Ph.D., Jeff Clark",
			want: Result{
				Value: "Jeff Clark Ph.D.",
			},
		},
		{
			input: "St Martin's Press",
			want: Result{
				Value: "St Martin's Press",
			},
		},
		{
			input: "Hackett, Sir John Winthrop",
			want: Result{
				Value: "Sir John Winthrop Hackett",
			},
		},
		{
			input: "Smith, K.C.",
			want: Result{
				Value:         "K.C. Smith",
				RequiresAudit: true,
				AuditReasons: []string{
					"[audit] King's Counsel OR initials - verify if suffix or given names: \"K.C.\"",
				},
			},
		},
		{
			input: "Denniston, J.D.",
			want: Result{
				Value:         "J.D. Denniston",
				RequiresAudit: true,
				AuditReasons: []string{
					"[audit] Juris Doctor OR initials - verify if suffix or given names: \"J.D.\"",
				},
			},
		},
		{
			input: "item, Be the first to review this",
			want: Result{
				Discarded: true,
				Literal:   true,
			},
		},
		{
			input: "Author, More by This",
			want: Result{
				Discarded: true,
				Literal:   true,
			},
		},
		{
			input: "AT&T Technology Systems",
			want: Result{
				Value:         "AT&T Technology Systems",
				RequiresAudit: true,
				AuditReasons: []string{
					"[collaboration] Name contains collaboration signifier: \"&\": \"AT&T Technology Systems\"",
				},
			},
		},
		{
			input: "John Grossman & Priscilla Dunhill",
			want: Result{
				Value:         "John Grossman & Priscilla Dunhill",
				RequiresAudit: true,
				AuditReasons: []string{
					"[collaboration] Name contains collaboration signifier: \"&\": \"John Grossman & Priscilla…\"",
				},
			},
		},
		{
			input: "S/orensen, Merethe Damsgaard",
			want: Result{
				Value:         "Merethe Damsgaard S/orensen",
				RequiresAudit: true,
				AuditReasons: []string{
					"[collaboration] Name contains collaboration signifier: \"/\": \"Merethe Damsgaard S/orens…\"",
				},
			},
		},
	}

	n, err := NewNames(realBooksConfigNames)
	require.NoError(t, err)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := n.Process(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Process(%q) failed\nGot:  %#v\nWant: %#v", tc.input, got, tc.want)
			}
		})
	}
}

// TestRealWorldSamples uses patterns identified in the user's export CSVs
// to verify the robustness of the normalization engines.
func TestRealWorldSamples(t *testing.T) {
	t.Parallel()

	t.Run("Punctuation bridge handling in Names", func(t *testing.T) {
		t.Parallel()
		n, err := NewNames(realBooksConfigNames)
		require.NoError(t, err)

		// Input: "Smith, John (ed.), Jr."
		// 1. Extract (ed.) -> "Smith, John, Jr." (Punctuation bridge logic)
		// 2. NaturalOrder -> "John Smith Jr."
		result := n.Process("Smith, John (ed.), Jr.")
		assert.Equal(t, "John Smith Jr.", result.Value)
		assert.Contains(t, result.Roles, "Editors")
	})

	t.Run("Video export markers", func(t *testing.T) {
		t.Parallel()
		config := MarkersConfig{
			Marker: map[string]string{
				"(Full)":   "Full Screen",
				"(MPAA)":   "MPAA Rated",
				"(IMDB)":   "IMDB Linked",
				"(Doctor)": "The Doctor",
			},
		}
		m, err := NewMarkers(config)
		require.NoError(t, err)

		// Complex title with multiple markers
		// "Doctor Who: The Caves of Androzani (Doctor) (Full) (MPAA)"
		result := m.Process("Doctor Who: The Caves of Androzani (Doctor) (Full) (MPAA)")
		assert.Equal(t, "Doctor Who: The Caves of Androzani", result.Value)
		assert.ElementsMatch(t, []string{"The Doctor", "Full Screen", "MPAA Rated"}, result.Markers)
	})

	t.Run("Metadata leakage detection", func(t *testing.T) {
		t.Parallel()
		config := NamesConfig{
			Audit: map[string]string{"Amazon.com": "[audit]", "Various": "[audit]"},
		}
		n, err := NewNames(config)
		require.NoError(t, err)

		// Placeholder names
		r1 := n.Process("Various")
		assert.True(t, r1.RequiresAudit)
		assert.Contains(t, r1.AuditReasons, "[audit]: \"Various\"")

		// Date leakage (will trigger IsUpper check or symbols)
		r2 := n.Process("-2014-12-04")
		assert.True(t, r2.RequiresAudit)
		// Should trigger [capital] (since it starts with a non-letter, non-digit character)
		assert.Condition(t, func() bool {
			for _, reason := range r2.AuditReasons {
				if strings.HasPrefix(reason, "[capital]") {
					return true
				}
			}
			return false
		}, "Expected [capital] audit reason")
	})

	t.Run("Parenthetical noise at end of string", func(t *testing.T) {
		t.Parallel()
		config := MarkersConfig{
			Marker: map[string]string{
				"(the ship's senile computer)": "Holly",
			},
		}
		m, err := NewMarkers(config)
		require.NoError(t, err)

		// "Red Dwarf: Holly (the ship's senile computer)"
		// Should remove the marker and the trailing space.
		result := m.Process("Red Dwarf: Holly (the ship's senile computer)")
		assert.Equal(t, "Red Dwarf: Holly", result.Value)
		assert.Contains(t, result.Markers, "Holly")
	})
}
