package isbn_test

import (
	"errors"
	"fmt"

	"github.com/DryHumour/readerware-to-tellico/isbn"
)

// ExampleNew_standard demonstrates parsing a perfectly clean ISBN.
func ExampleNew_standard() {
	parsed, err := isbn.New("9780306406157")
	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println(parsed.String())
	// Output: 9780306406157
}

// ExampleNew_readerware demonstrates how the parser strips whitespace and
// standard hyphens, but rejects alphabetic prefix tags.
func ExampleNew_readerware() {
	// Hyphens and spaces are fine — tags are not
	parsed, err := isbn.New("978-0-306-40615-7")
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println(parsed.String())

	_, tagErr := isbn.New("ISBN-13: 978-0-306-40615-7")
	if tagErr != nil {
		fmt.Println("Tag rejected")
	}
	// Output:
	// 9780306406157
	// Tag rejected
}

// ExampleNew_isbn10 demonstrates parsing an ISBN-10 ending in 'X'.
func ExampleNew_isbn10() {
	parsed, err := isbn.New("0-8044-2957-X")
	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println(parsed.String())
	// Output: 080442957X
}

// ExampleNew_forgiveness demonstrates the parser's "Forgiveness Policy".
// If the structure is perfect but the check-digit math fails, the parser
// returns the struct alongside the error so the caller can choose to save it.
func ExampleNew_forgiveness() {
	// 9780306406157 is valid. We intentionally change the last digit to 9.
	parsed, err := isbn.New("9780306406159")

	if errors.Is(err, isbn.ErrInvalidCheckDigit) {
		fmt.Println("Warning: Math failed, but structure saved.")
		fmt.Println("Recovered:", parsed.String())
	}

	// Output:
	// Warning: Math failed, but structure saved.
	// Recovered: 9780306406159
}

// ExampleNew_invalid demonstrates how the parser rejects structurally
// invalid strings, ensuring illegal states are unrepresentable.
func ExampleNew_invalid() {
	// Tag characters are invalid; rejected regardless of the digits that follow
	_, err := isbn.New("ISBN: 12345")
	if err != nil {
		fmt.Println("Rejected")
	}

	// 'X' is mathematically illegal in an ISBN-13
	_, err2 := isbn.New("978030640615X")
	if err2 != nil {
		fmt.Println("Rejected")
	}

	// Output:
	// Rejected
	// Rejected
}

// ExampleNew_typography demonstrates the parser safely stepping over
// invisible copy-paste artifacts like Bidi marks and Zero-Width Spaces.
func ExampleNew_typography() {
	// This string contains a Left-To-Right Mark (U+200E) and a Figure Dash (U+2012)
	messy := "978\u200E\u20120\u2012306\u201240615\u20127"

	parsed, err := isbn.New(messy)
	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println(parsed.String())
	// Output: 9780306406157
}
