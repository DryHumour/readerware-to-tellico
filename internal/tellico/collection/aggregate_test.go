package collection

import (
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/normalize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestNames(t testing.TB) *normalize.Names {
	t.Helper()
	names, err := normalize.NewNames(normalize.NamesConfig{
		Role: map[string]string{
			"(ed.)":    "Editors",
			"(illus.)": "Illustrators",
			"(trans.)": "Translators",
		},
	})
	require.NoError(t, err, "creating test Names config")
	return names
}

var testNameColumns = ColumnConfig{
	Names: map[string][]string{
		"Authors": {"AUTHOR", "AUTHOR2", "AUTHOR3", "AUTHOR4", "AUTHOR5", "AUTHOR6"},
		"Editors": {"EDITOR"},
	},
	Markers: map[string]bool{
		"TITLE":   true,
		"EDITION": true,
	},
	Categories: map[string]bool{
		"CATEGORY1": true,
		"CATEGORY2": true,
		"CATEGORY3": true,
	},
}

func TestAggregateNames(t *testing.T) {

	info := &collectionInfo{columns: testNameColumns}
	names := newTestNames(t)

	cases := []struct {
		name        string
		clean       map[string]string
		wantCredits map[string][]string
	}{
		{
			name:        "author in author column assigned to Authors",
			clean:       map[string]string{"AUTHOR": "Adams, Douglas"},
			wantCredits: map[string][]string{"Authors": {"Douglas Adams"}},
		},
		{
			name:        "author with no annotation in numbered column assigned to Authors",
			clean:       map[string]string{"AUTHOR3": "Adams, Douglas"},
			wantCredits: map[string][]string{"Authors": {"Douglas Adams"}},
		},
		{
			name:        "illustrator annotation in author column cross-assigned to Illustrators",
			clean:       map[string]string{"AUTHOR": "Adams, Douglas (illus.)"},
			wantCredits: map[string][]string{"Illustrators": {"Douglas Adams"}},
		},
		{
			name:  "mixed: plain author and editor annotation in same cell",
			clean: map[string]string{"AUTHOR": "Adams, Douglas; Jones, Bob (ed.)"},
			wantCredits: map[string][]string{
				"Authors": {"Douglas Adams"},
				"Editors": {"Bob Jones"},
			},
		},
		{
			name:        "editor annotation in editor column assigned to Editors",
			clean:       map[string]string{"EDITOR": "Adams, Douglas (ed.)"},
			wantCredits: map[string][]string{"Editors": {"Douglas Adams"}},
		},
		{
			name:  "cross-role: translator annotation appended after primary authors",
			clean: map[string]string{"AUTHOR": "Smith, John", "AUTHOR2": "Doe, Jane (trans.)"},
			wantCredits: map[string][]string{
				"Authors":     {"John Smith"},
				"Translators": {"Jane Doe"},
			},
		},
		{
			name:        "deduplication: same name in two columns appears once",
			clean:       map[string]string{"AUTHOR": "Adams, Douglas", "AUTHOR2": "Adams, Douglas"},
			wantCredits: map[string][]string{"Authors": {"Douglas Adams"}},
		},
		{
			name:  "squeeze-style: space before semicolon handled correctly",
			clean: map[string]string{"AUTHOR": "Adams, Douglas ; Jones, Bob (ed.)"},
			wantCredits: map[string][]string{
				"Authors": {"Douglas Adams"},
				"Editors": {"Bob Jones"},
			},
		},
		{
			name: "preserves column order: names from AUTHOR, AUTHOR2, AUTHOR3 in that order",
			clean: map[string]string{
				"AUTHOR":  "First Author",
				"AUTHOR2": "Second Author",
				"AUTHOR3": "Third Author",
			},
			wantCredits: map[string][]string{
				"Authors": {"First Author", "Second Author", "Third Author"},
			},
		},
		{
			name: "cross-role names appended after primary names in column order",
			clean: map[string]string{
				"AUTHOR":  "Primary One",
				"AUTHOR2": "Primary Two (trans.)",
				"AUTHOR3": "Primary Three",
			},
			wantCredits: map[string][]string{
				"Authors":     {"Primary One", "Primary Three"},
				"Translators": {"Primary Two"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ent, err := newBasicEntry(info, tc.clean, images.Row{})
			require.NoError(t, err)

			AggregateNames(ent, names)

			assert.Equal(t, tc.wantCredits, ent.agg.credits)
		})
	}
}

func TestAggregateNames_NilNames(t *testing.T) {

	info := &collectionInfo{columns: testNameColumns}
	ent, err := newBasicEntry(info, map[string]string{"AUTHOR": "Smith, John"}, images.Row{})
	require.NoError(t, err)
	AggregateNames(ent, nil)
	assert.Empty(t, ent.agg.credits, "nil Names should produce no credits")
}

func TestAggregateMarkers(t *testing.T) {

	markers, err := normalize.NewMarkers(normalize.MarkersConfig{
		Marker: map[string]string{
			"(signed)":       "<signed>",
			"(out of print)": "<out_of_print>",
		},
	})
	require.NoError(t, err, "creating test Markers config")

	info := &collectionInfo{columns: testNameColumns}

	cases := []struct {
		name        string
		clean       map[string]string
		wantMarkers map[string]bool
	}{
		{
			name:        "detects marker in TITLE column",
			clean:       map[string]string{"TITLE": "A Book (signed)"},
			wantMarkers: map[string]bool{"<signed>": true},
		},
		{
			name:        "detects marker in EDITION column",
			clean:       map[string]string{"EDITION": "First (out of print)"},
			wantMarkers: map[string]bool{"<out_of_print>": true},
		},
		{
			name:        "no markers when none present",
			clean:       map[string]string{"TITLE": "A Plain Book"},
			wantMarkers: map[string]bool{},
		},
		{
			name:        "detects markers in multiple columns simultaneously",
			clean:       map[string]string{"TITLE": "A Book (signed)", "EDITION": "First (out of print)"},
			wantMarkers: map[string]bool{"<signed>": true, "<out_of_print>": true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ent, err := newBasicEntry(info, tc.clean, images.Row{})
			require.NoError(t, err)

			AggregateMarkers(ent, markers)

			assert.Equal(t, tc.wantMarkers, ent.agg.markers)
		})
	}
}

func TestAggregateMarkers_NilMarkers(t *testing.T) {
	info := &collectionInfo{columns: testNameColumns}
	ent, err := newBasicEntry(info, map[string]string{"TITLE": "A Book (signed)"}, images.Row{})
	require.NoError(t, err)
	AggregateMarkers(ent, nil)
	assert.Empty(t, ent.agg.markers, "nil Markers should produce no markers")
}

func TestAggregation_Credits(t *testing.T) {

	agg := newAggregation()
	agg.credits = map[string][]string{
		"Authors": {"John Doe", "Jane Smith"},
	}

	t.Run("returns names with abbreviation", func(t *testing.T) {
		result := agg.Credits("Authors", "[ed.]")
		assert.Equal(t, []string{"John Doe [ed.]", "Jane Smith [ed.]"}, result)
	})

	t.Run("returns names without abbreviation", func(t *testing.T) {
		result := agg.Credits("Authors", "")
		assert.Equal(t, []string{"John Doe", "Jane Smith"}, result)
	})

	t.Run("returns nil for empty role", func(t *testing.T) {
		result := agg.Credits("Editors", "[ed.]")
		assert.Nil(t, result)
	})
}

func TestAggregation_AddCredit(t *testing.T) {

	agg := newAggregation()

	t.Run("adds name to new role", func(t *testing.T) {
		agg.AddCredit("Authors", "John Doe")
		assert.Equal(t, []string{"John Doe"}, agg.credits["Authors"])
	})

	t.Run("deduplicates within role", func(t *testing.T) {
		agg2 := newAggregation()
		agg2.AddCredit("Authors", "Jane Smith")
		agg2.AddCredit("Authors", "Jane Smith")
		assert.Equal(t, []string{"Jane Smith"}, agg2.credits["Authors"])
	})

	t.Run("preserves order of additions", func(t *testing.T) {
		agg2 := newAggregation()
		agg2.AddCredit("Authors", "First")
		agg2.AddCredit("Authors", "Second")
		agg2.AddCredit("Authors", "Third")
		assert.Equal(t, []string{"First", "Second", "Third"}, agg2.credits["Authors"])
	})
}

func TestAggregation_HasMarker(t *testing.T) {

	agg := newAggregation()
	agg.markers = map[string]bool{
		"<signed>": true,
	}

	t.Run("returns true for existing marker", func(t *testing.T) {
		assert.True(t, agg.HasMarker("<signed>"))
	})

	t.Run("returns false for non-existent marker", func(t *testing.T) {
		assert.False(t, agg.HasMarker("<out_of_print>"))
	})
}
