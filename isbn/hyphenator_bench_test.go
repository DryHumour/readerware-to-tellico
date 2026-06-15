package isbn

import (
	"bufio"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkHyphenator_Hyphenate(b *testing.B) {
	ranges, err := ParseISBNRangesJSON(isbnRangesJSON)
	require.NoError(b, err)

	h, err := NewHyphenator(ranges)
	require.NoError(b, err)

	file, err := os.Open("testdata/isbn.txt")
	require.NoError(b, err)
	defer file.Close()

	var isbns []ISBN
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		i, err := New(scanner.Text())
		if err != nil {
			continue
		}
		isbns = append(isbns, i)
	}
	require.NoError(b, scanner.Err())

	if len(isbns) == 0 {
		b.Fatal("no ISBNs found in testdata/isbn.txt")
	}

	b.ResetTimer()
	for b.Loop() {
		for _, isbn := range isbns {
			_, _ = h.Hyphenate(isbn)
		}
	}
}
