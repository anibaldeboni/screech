package components

import (
	"reflect"
	"testing"

	"github.com/veandco/go-sdl2/sdl"
)

func TestParseLines(t *testing.T) {
	// Initialize a TextView instance with a specific width
	textView := NewTextView(nil, TextViewSize{Width: 10, Height: 5}, sdl.Point{X: 0, Y: 0})

	// Define input text that exceeds the width
	input := []string{
		"This is a long line that should be wrapped.",
		"Short",
		"Another long line that needs wrapping.",
	}

	// Expected output after parsing
	expected := []string{
		"This is a",
		"long line",
		"that",
		"should be",
		"wrapped.",
		"Short",
		"Another",
		"long line",
		"that",
		"needs",
		"wrapping.",
	}

	// Call the parseLines method with the input text
	output := textView.parseLines(input)

	// Verify that the output matches the expected result
	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Expected %v, but got %v", expected, output)
	}
}
