package normalize

import (
	"testing"
)

// FuzzProcessNames targets the entire Names pipeline.
func FuzzProcessNames(f *testing.F) {
	f.Add("Adams, Douglas (ed.)")
	f.Add("Prince")
	f.Add("n/a")
	f.Add("Doe, John, Jr.")

	config := NamesConfig{
		Keep:       []string{"Prince"},
		Discard:    []string{"n/a"},
		Scrub:      []string{"noise"},
		Role:       map[string]string{"(ed.)": "Editors"},
		Suffixes:   []string{"Jr.", "Sr."},
		Honorifics: []string{"Sir", "Dr."},
		Corporate:  []string{"Orchestra"},
	}
	n, err := NewNames(config)
	if err != nil {
		f.Fatalf("failed to create Names: %v", err)
	}

	f.Fuzz(func(t *testing.T, input string) {
		_ = n.Process(input)
	})
}

// FuzzNaturalOrder targets the complex comma-separated name flipping logic.
func FuzzNaturalOrder(f *testing.F) {
	f.Add("Smith, John")
	f.Add("Doe, John, Jr.")
	f.Add("Redgrave, Michael, Sir")
	f.Add("London Orchestra, The")
	f.Add("Smith,,John")
	f.Add("Smith, John,")
	f.Add("Single")
	f.Add("A,B,C,D,E")
	f.Add("")
	f.Add("   ")
	f.Add("Müller, Hans")
	f.Add("東京, 東京")

	config := NamesConfig{
		Suffixes:   []string{"Jr.", "Sr.", "Ph.D."},
		Honorifics: []string{"Sir", "Dr.", "Prof."},
		Corporate:  []string{"Orchestra", "University", "Dept."},
	}
	n, err := NewNames(config)
	if err != nil {
		f.Fatalf("failed to create Names: %v", err)
	}

	f.Fuzz(func(t *testing.T, input string) {
		r := n.Begin(input)
		_ = n.NaturalOrder(r)
	})
}
