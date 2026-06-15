package strutil

import (
	"regexp"
	"testing"
)

func FuzzIsArtifact(f *testing.F) {
	f.Add("#productDescription { font-family: verdana; }")
	f.Add("Normal Book Title")
	f.Add("Author, John")
	f.Add("color: #333; background: #fff;")
	f.Add("#id-selector")
	f.Add("Doctor Who: The Complete Tenth Series")
	f.Add("Travel : Canada : Ontario : Ottawa")
	f.Add("Director/Supervising Producer Martin Wood")
	f.Add("Quick & Easy Quiltmaking: Twenty-Six Projects")
	f.Add("Te Kanawa/Studer/Araiza/Ramey")
	f.Add("It Might as Well Be Spring")
	// Unicode edge cases (realistic for CSV data)
	f.Add("Müller GmbH & Co. KG")
	f.Add("José Saramago")
	f.Add("東京")
	f.Add("Москва")
	f.Add("")       // empty string
	f.Add("   ")    // only spaces
	f.Add("\n\n\n") // only newlines
	f.Add("\t\t\t") // only tabs

	f.Fuzz(func(t *testing.T, input string) {
		// Ensure no panics
		_ = IsArtifact(input)
	})
}

func FuzzSqueeze(f *testing.F) {
	f.Add("hello   world")
	f.Add("hello\t\tworld")
	f.Add("  hello world  ")
	f.Add("hello world")
	f.Add("")
	f.Add("   \t  ")
	f.Add("helloworld")
	f.Add("  \t  hello world")
	f.Add("hello world  \t  ")
	f.Add("multiple   spaces   here")
	f.Add("tabs\tand\tspaces")
	f.Add("\n\n\n")
	f.Add("  \n  ")
	// Unicode edge cases (realistic for CSV data)
	f.Add("Müller   GmbH")
	f.Add("東京　大阪") // fullwidth spaces
	f.Add("café  au  lait")
	f.Add("naïve   ïn   wïntër")
	f.Add("") // empty string

	f.Fuzz(func(t *testing.T, input string) {
		// Ensure no panics
		_ = Squeeze(input)
		_ = SqueezePreserveNewlines(input)
	})
}

func FuzzEngineRegexSqueeze(f *testing.F) {
	re := regexp.MustCompile(`\d+`)

	f.Add("abc123def")
	f.Add("a1b2c3")
	f.Add("abc def")
	f.Add("  123  abc 456  ")
	f.Add("")
	f.Add("123abc")
	f.Add("abc123")
	f.Add("123456")
	f.Add("no numbers here")
	f.Add("123")
	f.Add("abc")
	f.Add("1 2 3 4 5")
	f.Add("\n123\n456\n")
	f.Add("  \n  ")
	// Unicode edge cases (realistic for CSV data)
	f.Add("abc１２３def") // fullwidth digits
	f.Add("Müller123GmbH")
	f.Add("café456au789lait")
	f.Add("") // empty string

	f.Fuzz(func(t *testing.T, input string) {
		// Ensure no panics for all variants
		_, _ = ExtractAndSqueeze(input, re)
		_, _ = ExtractAndSqueezePreserveNewlines(input, re)
		_ = ReplaceAndSqueeze(input, re)
		_ = ReplaceAndSqueezePreserveNewlines(input, re)
	})
}
