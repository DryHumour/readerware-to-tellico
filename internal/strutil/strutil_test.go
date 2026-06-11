package strutil

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckbox(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"true lowercase", "true", "true"},
		{"true uppercase", "TRUE", "true"},
		{"true mixed case", "TrUe", "true"},
		{"t", "t", "true"},
		{"T", "T", "true"},
		{"yes", "yes", "true"},
		{"YES", "YES", "true"},
		{"y", "y", "true"},
		{"Y", "Y", "true"},
		{"1", "1", "true"},
		{"false", "false", ""},
		{"no", "no", ""},
		{"0", "0", ""},
		{"empty string", "", ""},
		{"random text", "something", ""},
		{"true with whitespace", "  true  ", "true"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := Checkbox(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDimensions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple x", "6 x 9", "6×9"},
		{"multiple spaces", "6  x  9", "6×9"},
		{"no x", "6 9", "6 9"},
		{"uppercase X", "6 X 9", "6 X 9"},
		{"three dimensions", "6 x 9 x 2", "6×9×2"},
		{"decimal values", "5.5 x 8.5", "5.5×8.5"},
		{"empty string", "", ""},
		{"only x", "x", "×"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := Dimensions(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHTMLToText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		contains string
	}{
		{"simple HTML", "<p>Hello</p>", "Hello"},
		{"bold", "<b>text</b>", "text"},
		{"link", "<a href='#'>link</a>", "link"},
		{"empty string", "", ""},
		{"plain text", "Hello world", "Hello world"},
		{"HTML with whitespace", "<p>  Hello  </p>", "Hello"},
		{"random closing tag without opening", "Some text</li> more text", "Some text"},
		{"multiple random closing tags", "text</li></p></div>", "text"},
		{"unclosed opening tag", "<p>Some text", "Some text"},
		{"nested tags", "<div><p>text</p></div>", "text"},
		{"tags with attributes", "<a href='http://example.com' class='link'>text</a>", "text"},
		{"script tag", "<script>alert('xss')</script>text", "text"},
		{"style tag", "<style>body{color:red}</style>text", "text"},
		{"malformed HTML with garbage", "text <broken> more </broken> end", "text"},
		{"br tag", "line1<br>line2", "line1"},
		{"hr tag", "above<hr>below", "above"},
		{"list items", "<ul><li>item1</li><li>item2</li></ul>", "item1"},
		{"table", "<table><tr><td>cell</td></tr></table>", "cell"},
		{"mixed valid and invalid", "<p>valid</p></li>invalid", "valid"},
		{"empty tags", "<p></p><b></b>", "**"},    // html2text converts empty tags to **
		{"HTML entities", "&lt;&gt;&amp;", "<>&"}, // html2text decodes entities
		{"HTML comments", "<!-- comment -->text", "text"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := HTMLToText(tc.input)
			if tc.contains != "" {
				assert.Contains(t, result, tc.contains)
			} else {
				assert.Equal(t, tc.input, result)
			}
		})
	}
}

func TestHyphenateISBN(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"10-digit ISBN", "0306406152", "0-306-40615-2"},
		{"13-digit ISBN", "9780306406157", "978-0-306-40615-7"},
		{"already hyphenated", "0-306-40615-2", "0-306-40615-2"},
		{"empty string", "", ""},
		{"invalid ISBN", "not-an-isbn", "not-an-isbn"}, // Returns input as-is
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := HyphenateISBN(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestKeywords(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"comma separated", "fiction,science", "fiction;science"},
		{"semicolon separated", "fiction;science", "fiction;science"},
		{"slash separated", "fiction/science", "fiction;science"},
		{"mixed separators", "fiction,science/history", "fiction;science;history"},
		{"single keyword", "fiction", "fiction"},
		{"empty string", "", ""},
		{"whitespace around separators", " fiction , science ", "fiction;science"},
		{"multiple commas", "a,b,c", "a;b;c"},
		{"trailing separator", "fiction,", "fiction"},
		{"leading separator", ",fiction", "fiction"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := Keywords(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPrice(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"non-zero price", "19.99", "19.99"},
		{"zero price", "0.00", ""},
		{"zero with whitespace", "  0.00  ", ""},
		{"empty string", "", ""},
		{"integer price", "10", "10"},
		{"large price", "999.99", "999.99"},
		{"price with currency symbol", "$19.99", "$19.99"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := Price(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRating1to5(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"integer 1", "1", "1"},
		{"integer 5", "5", "5"},
		{"integer 3", "3", "3"},
		{"decimal 4.5", "4.5", "5"},
		{"decimal 4.4", "4.4", "4"},
		{"decimal 1.1", "1.1", "1"},
		{"decimal 1.5", "1.5", "2"},
		{"below range", "0", ""},
		{"above range", "6", ""},
		{"decimal below range", "0.5", ""},
		{"decimal above range", "5.5", ""},
		{"rating in text", "Rating: 4 stars", "4"},
		{"empty string", "", ""},
		{"non-numeric", "bad", ""},
		{"multiple numbers", "4.5 out of 5", "5"}, // Extracts 4.5 which rounds to 5
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := Rating1to5(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSqueeze(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"multiple spaces", "hello   world", "hello world"},
		{"tabs", "hello\tworld", "hello world"},
		{"mixed whitespace", "hello  \t  world", "hello world"},
		{"leading/trailing spaces", "  hello world  ", "hello world"},
		{"single space", "hello world", "hello world"},
		{"empty string", "", ""},
		{"only whitespace", "   \t  ", ""},
		{"no whitespace", "helloworld", "helloworld"},
		{"leading mixed whitespace", "  \t  hello world", "hello world"},
		{"trailing mixed whitespace", "hello world  \t  ", "hello world"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := Squeeze(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestXMLEscape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		contains string
	}{
		{"less than", "<tag>", "&lt;"},
		{"greater than", "value>", "&gt;"},
		{"ampersand", "a&b", "&amp;"},
		{"quote", "\"text\"", "&#34;"},
		{"apostrophe", "'text'", "&#39;"},
		{"mixed special chars", "<>&\"'", "&lt;&gt;&amp;&#34;&#39;"},
		{"plain text", "hello", "hello"},
		{"empty string", "", ""},
		{"XSS attempt", "<script>alert('xss')</script>", "&lt;"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := XMLEscape(tc.input)
			assert.Contains(t, result, tc.contains)
		})
	}
}

func TestContainsHTML(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple tag", "<p>text</p>", true},
		{"self-closing tag", "<br/>", true},
		{"tag with attributes", "<a href='#'>link</a>", true},
		{"HTML entity", "&lt;", true},
		{"numeric entity", "&#60;", true},
		{"hex entity", "&#x3C;", true},
		{"plain text", "hello world", false},
		{"empty string", "", false},
		{"angle bracket without letter", "<>", false},
		{"ampersand without semicolon", "a & b", false},
		{"mixed case tag", "<DIV>text</DIV>", true},
		{"closing tag only", "text</p>", true},
		{"multiple tags", "<div><p>text</p></div>", true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := ContainsHTML(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractAndSqueeze(t *testing.T) {
	t.Parallel()

	re := regexp.MustCompile(`\d+`)

	cases := []struct {
		name          string
		input         string
		expectedClean string
		expectedMatch []string
	}{
		{"simple extraction", "abc123def", "abc def", []string{"123"}},
		{"multiple matches", "a1b2c3", "a b c", []string{"1", "2", "3"}},
		{"no match", "abc def", "abc def", nil},
		{"match with spaces", "  123  abc 456  ", "abc", []string{"123", "456"}},
		{"empty string", "", "", nil},
		{"leading match", "123abc", "abc", []string{"123"}},
		{"trailing match", "abc123", "abc", []string{"123"}},
		{"consecutive matches", "123456", "", []string{"123456"}},
		{"match in middle", "abc123def", "abc def", []string{"123"}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			clean, matches := ExtractAndSqueeze(tc.input, re)
			assert.Equal(t, tc.expectedClean, clean)
			assert.Equal(t, tc.expectedMatch, matches)
		})
	}
}

func TestExtractAndSqueezePreserveNewlines(t *testing.T) {
	t.Parallel()

	re := regexp.MustCompile(`\d+`)

	cases := []struct {
		name          string
		input         string
		expectedClean string
		expectedMatch []string
	}{
		{"simple extraction", "abc123def", "abc def", []string{"123"}},
		{"preserve single newline", "abc\n123def", "abc\ndef", []string{"123"}},
		{"preserve multiple newlines", "abc\n\n123def", "abc\n\ndef", []string{"123"}},
		{"newlines around match", "abc\n123\ndef", "abc\n\ndef", []string{"123"}},
		{"no match with newlines", "abc\ndef\nghi", "abc\ndef\nghi", nil},
		{"empty string", "", "", nil},
		{"whitespace and newlines", "  \n  abc  \n  ", "abc", nil},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			clean, matches := ExtractAndSqueezePreserveNewlines(tc.input, re)
			assert.Equal(t, tc.expectedClean, clean)
			assert.Equal(t, tc.expectedMatch, matches)
		})
	}
}

func TestJoinParts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		parts    []string
		expected string
	}{
		{"simple join", []string{"hello", "world"}, "hello world"},
		{"with empty string", []string{"hello", "", "world"}, "hello world"},
		{"all empty", []string{"", "", ""}, ""},
		{"single part", []string{"hello"}, "hello"},
		{"no parts", []string{}, ""},
		{"multiple empties", []string{"", "hello", "", "world", ""}, "hello world"},
		{"nil slice", nil, ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := JoinParts(tc.parts...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParagraphs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"single newline", "line1\nline2", "line1<br/>line2"},
		{"multiple newlines", "line1\n\nline2", "line1<br/><br/>line2"},
		{"no newlines", "single line", "single line"},
		{"empty string", "", ""},
		{"trailing newline", "line1\n", "line1<br/>"},
		{"leading newline", "\nline1", "<br/>line1"},
		{"mixed content", "para1\npara2\npara3", "para1<br/>para2<br/>para3"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := Paragraphs(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestReplaceAndSqueeze(t *testing.T) {
	t.Parallel()

	re := regexp.MustCompile(`\d+`)

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple replacement", "abc123def", "abc def"},
		{"multiple replacements", "a1b2c3", "a b c"},
		{"no match", "abc def", "abc def"},
		{"match with spaces", "  123  abc 456  ", "abc"},
		{"empty string", "", ""},
		{"leading match", "123abc", "abc"},
		{"trailing match", "abc123", "abc"},
		{"consecutive matches", "123456", ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := ReplaceAndSqueeze(tc.input, re)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestReplaceAndSqueezePreserveNewlines(t *testing.T) {
	t.Parallel()

	re := regexp.MustCompile(`\d+`)

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple replacement", "abc123def", "abc def"},
		{"preserve single newline", "abc\n123def", "abc\ndef"},
		{"preserve multiple newlines", "abc\n\n123def", "abc\n\ndef"},
		{"newlines around match", "abc\n123\ndef", "abc\n\ndef"},
		{"no match with newlines", "abc\ndef\nghi", "abc\ndef\nghi"},
		{"empty string", "", ""},
		{"whitespace and newlines", "  \n  abc  \n  ", "abc"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := ReplaceAndSqueezePreserveNewlines(tc.input, re)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSplitList(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected []string
	}{
		{"semicolon separated", "item1;item2;item3", []string{"item1", "item2", "item3"}},
		{"slash separated", "item1/item2/item3", []string{"item1", "item2", "item3"}},
		{"mixed separators", "item1;item2/item3", []string{"item1", "item2", "item3"}},
		{"with whitespace", " item1 ; item2 / item3 ", []string{"item1", "item2", "item3"}},
		{"empty string", "", nil},
		{"single item", "item1", []string{"item1"}},
		{"trailing separator", "item1;item2;", []string{"item1", "item2"}},
		{"leading separator", ";item1;item2", []string{"item1", "item2"}},
		{"empty parts", "item1;;item2", []string{"item1", "item2"}},
		{"commas not separators", "Smith, John; Doe, Jane", []string{"Smith, John", "Doe, Jane"}},
		{"multiple slashes", "a/b/c/d", []string{"a", "b", "c", "d"}},
		{"N/A preserved", "N/A", []string{"N/A"}},
		{"N/A with other items", "item1;N/A;item2", []string{"item1", "N/A", "item2"}},
		{"N/A at start with other items", "N/A;item1;item2", []string{"N/A", "item1", "item2"}},
		{"N/A at end with other items", "item1/item2/N/A", []string{"item1", "item2", "N/A"}},
		{"n/a lowercase preserved", "n/a", []string{"n/a"}},
		{"N/A mixed case preserved", "N/a", []string{"N/a"}},
		{"N/A not split on slash", "N/A", []string{"N/A"}},
		{"N/ not special", "N/", []string{"N"}},
		{"/A not special", "/A", []string{"A"}},
		{"N/ seq not special", "N/B/C", []string{"N", "B", "C"}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := SplitList(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSqueezePreserveNewlines(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"preserve single newline", "hello\nworld", "hello\nworld"},
		{"preserve multiple newlines", "hello\n\nworld", "hello\n\nworld"},
		{"collapse horizontal ws", "hello   world", "hello world"},
		{"collapse tabs", "hello\t\tworld", "hello world"},
		{"mixed ws with newline", "hello  \n  world", "hello\nworld"},
		{"leading newlines", "\n\nhello", "hello"},
		{"trailing newlines", "hello\n\n", "hello"},
		{"only newlines", "\n\n\n", ""},
		{"empty string", "", ""},
		{"no whitespace", "helloworld", "helloworld"},
		{"newlines with spaces", "hello \n \n world", "hello\n\nworld"},
		{"single space preserved", "hello world", "hello world"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := SqueezePreserveNewlines(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestToStringSlice(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    any
		expected []string
		wantErr  bool
	}{
		{"[]string input", []string{"a", "b", "c"}, []string{"a", "b", "c"}, false},
		{"[]any with strings", []any{"a", "b", "c"}, []string{"a", "b", "c"}, false},
		{"[]any with mixed types", []any{"a", 123, true}, []string{"a", "123", "true"}, false},
		{"[]any with nil", []any{nil, "a", nil}, []string{"<nil>", "a", "<nil>"}, false},
		{"nil input", nil, nil, false},
		{"empty []string", []string{}, []string{}, false},
		{"empty []any", []any{}, []string{}, false},
		{"invalid type - string", "not a slice", nil, true},
		{"invalid type - int", 123, nil, true},
		{"invalid type - map", map[string]string{}, nil, true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := ToStringSlice(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestToStringStringMap(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    any
		expected map[string]string
		wantErr  bool
	}{
		{"map[string]string input", map[string]string{"a": "1", "b": "2"}, map[string]string{"a": "1", "b": "2"}, false},
		{"map[string]any with strings", map[string]any{"a": "1", "b": "2"}, map[string]string{"a": "1", "b": "2"}, false},
		{"map[string]any with mixed types", map[string]any{"a": 123, "b": true}, map[string]string{"a": "123", "b": "true"}, false},
		{"map[string]any with nil values", map[string]any{"a": nil, "b": "2"}, map[string]string{"a": "<nil>", "b": "2"}, false},
		{"nil input", nil, nil, false},
		{"empty map[string]string", map[string]string{}, map[string]string{}, false},
		{"empty map[string]any", map[string]any{}, map[string]string{}, false},
		{"invalid type - string", "not a map", nil, true},
		{"invalid type - int", 123, nil, true},
		{"invalid type - slice", []string{}, nil, true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := ToStringStringMap(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
