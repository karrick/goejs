package goejs_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/karrick/goejs"
)

func ensureError(tb testing.TB, testCase string, err error, contains string) {
	if err == nil || !strings.Contains(err.Error(), contains) {
		tb.Errorf("Case: %#q; Actual: %v; Expected: %s", testCase, err, contains)
	}
}

func ensureBad(tb testing.TB, input, errorMessage, remainder string) {
	_, buf2, err := goejs.DecodedStringFromJSON([]byte(input))
	ensureError(tb, input, err, errorMessage)
	if actual, expected := string(buf2), remainder; actual != expected {
		tb.Errorf("Input: %#q; Remainder Actual: %#q; Expected: %#q; Error: %s", input, actual, expected, err)
	}
}

func ensureGood(tb testing.TB, input, expected string) {
	buf := goejs.EncodedJSONFromString(nil, input)
	if actual := string(buf); actual != expected {
		tb.Errorf("Input: %#q; Actual: %#q; Expected: %#q", input, actual, expected)
	}
	output, buf2, err := goejs.DecodedStringFromJSON([]byte(expected))
	if err != nil {
		tb.Errorf("Input: %#q: %s", input, err)
	}
	if input != output {
		tb.Errorf("Input: %#q; Output: %#q", input, output)
	}
	if actual, expected := string(buf2), ""; actual != expected {
		tb.Errorf("Input: %#q; Remainder Actual: %#q; Expected: %#q", input, actual, expected)
	}
}

////////////////////////////////////////

func ExampleDecode() {
	decoded, remainder, err := goejs.DecodedStringFromJSON([]byte("\"\\u0001\\u2318 a\" some extra bytes after final quote"))
	if err != nil {
		fmt.Println(err)
	}
	if actual, expected := string(remainder), " some extra bytes after final quote"; actual != expected {
		fmt.Printf("Remainder Actual: %#q; Expected: %#q\n", actual, expected)
	}
	fmt.Printf("%v", decoded)
	// Output: ⌘ a
}

func ExampleEncode() {
	encoded := goejs.EncodedJSONFromString([]byte("prefix:"), "⌘ a")
	fmt.Printf("%s", encoded)
	// Output: prefix:"\u0001\u2318 a"
}

func TestString(t *testing.T) {
	ensureBad(t, `"`, "short buffer", "\"")
	ensureBad(t, `..`, "expected initial '\"'", "..")
	ensureBad(t, `".`, "expected final '\"'", "\".")

	ensureGood(t, "", "\"\"")
	ensureGood(t, "a", "\"a\"")
	ensureGood(t, "ab", "\"ab\"")
	ensureGood(t, "a\"b", "\"a\\\"b\"")
	ensureGood(t, "a\\b", "\"a\\\\b\"")
	ensureGood(t, "a/b", "\"a\\/b\"")

	ensureGood(t, "a\bb", `"a\bb"`)
	ensureGood(t, "a\fb", `"a\fb"`)
	ensureGood(t, "a\nb", `"a\nb"`)
	ensureGood(t, "a\rb", `"a\rb"`)
	ensureGood(t, "a\tb", `"a\tb"`)
	ensureGood(t, "a	b", `"a\tb"`) // tab byte between a and b

	ensureBad(t, "\"\\u\"", "short buffer", "\"")
	ensureBad(t, "\"\\u.\"", "short buffer", ".\"")
	ensureBad(t, "\"\\u..\"", "short buffer", "..\"")
	ensureBad(t, "\"\\u...\"", "short buffer", "...\"")

	ensureBad(t, "\"\\u////\"", "invalid byte", "////\"") // < '0'
	ensureBad(t, "\"\\u::::\"", "invalid byte", "::::\"") // > '9'
	ensureBad(t, "\"\\u@@@@\"", "invalid byte", "@@@@\"") // < 'A'
	ensureBad(t, "\"\\uGGGG\"", "invalid byte", "GGGG\"") // > 'F'
	ensureBad(t, "\"\\u````\"", "invalid byte", "````\"") // < 'a'
	ensureBad(t, "\"\\ugggg\"", "invalid byte", "gggg\"") // > 'f'

	ensureGood(t, "⌘ ", "\"\\u0001\\u2318 \"")
	ensureGood(t, "😂 ", "\"\\u0001\\uD83D\\uDE02 \"")

	ensureBad(t, "\"\\uD83D\"", "surrogate pair", "")
	ensureBad(t, "\"\\uD83D\\u\"", "surrogate pair", "u\"")
	ensureBad(t, "\"\\uD83D\\uD\"", "surrogate pair", "uD\"")
	ensureBad(t, "\"\\uD83D\\uDE\"", "surrogate pair", "uDE\"")
	ensureBad(t, "\"\\uD83D\\uDE0\"", "invalid byte", "uDE0\"")
}
