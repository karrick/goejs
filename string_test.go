package goejs_test

import (
	"fmt"
	"testing"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/karrick/goejs"
)

func stringEnsureBad(tb testing.TB, input, errorMessage, remainder string) {
	_, buf2, err := goejs.DecodedStringFromJSON([]byte(input))
	ensureError(tb, input, err, errorMessage)
	if actual, expected := string(buf2), remainder; actual != expected {
		tb.Errorf("Input: %#q; Remainder Actual: %#q; Expected: %#q; Error: %s", input, actual, expected, err)
	}
}

func stringEnsureGood(tb testing.TB, input, expected string) {
	buf := goejs.AppendEncodedJSONFromString(nil, input)
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

func ExampleStringDecode() {
	decoded, remainder, err := goejs.DecodedStringFromJSON([]byte("\"\\u0001\\u2318 a\" some extra bytes after final quote"))
	if err != nil {
		fmt.Println(err)
	}
	if actual, expected := string(remainder), " some extra bytes after final quote"; actual != expected {
		fmt.Printf("Remainder Actual: %#q; Expected: %#q\n", actual, expected)
	}
	fmt.Printf("%#q", decoded)
	// Output: "\x01⌘ a"
}

func ExampleStringEncode() {
	encoded := goejs.AppendEncodedJSONFromString([]byte("prefix:"), "⌘ a")
	fmt.Printf("%s", encoded)
	// Output: prefix:"\u0001\u2318 a"
}

func TestString(t *testing.T) {
	stringEnsureBad(t, `"`, "short buffer", "\"")
	stringEnsureBad(t, `..`, "expected initial '\"'", "..")
	stringEnsureBad(t, `".`, "expected final '\"'", "\".")

	stringEnsureGood(t, "", "\"\"")
	stringEnsureGood(t, "a", "\"a\"")
	stringEnsureGood(t, "ab", "\"ab\"")
	stringEnsureGood(t, "a\"b", "\"a\\\"b\"")
	stringEnsureGood(t, "a\\b", "\"a\\\\b\"")
	stringEnsureGood(t, "a/b", "\"a\\/b\"")

	stringEnsureGood(t, "a\bb", `"a\bb"`)
	stringEnsureGood(t, "a\fb", `"a\fb"`)
	stringEnsureGood(t, "a\nb", `"a\nb"`)
	stringEnsureGood(t, "a\rb", `"a\rb"`)
	stringEnsureGood(t, "a\tb", `"a\tb"`)
	stringEnsureGood(t, "a	b", `"a\tb"`) // tab byte between a and b

	stringEnsureBad(t, "\"\\u\"", "short buffer", "\"")
	stringEnsureBad(t, "\"\\u.\"", "short buffer", ".\"")
	stringEnsureBad(t, "\"\\u..\"", "short buffer", "..\"")
	stringEnsureBad(t, "\"\\u...\"", "short buffer", "...\"")

	stringEnsureBad(t, "\"\\u////\"", "invalid byte", "////\"") // < '0'
	stringEnsureBad(t, "\"\\u::::\"", "invalid byte", "::::\"") // > '9'
	stringEnsureBad(t, "\"\\u@@@@\"", "invalid byte", "@@@@\"") // < 'A'
	stringEnsureBad(t, "\"\\uGGGG\"", "invalid byte", "GGGG\"") // > 'F'
	stringEnsureBad(t, "\"\\u````\"", "invalid byte", "````\"") // < 'a'
	stringEnsureBad(t, "\"\\ugggg\"", "invalid byte", "gggg\"") // > 'f'

	stringEnsureGood(t, "⌘ ", "\"\\u0001\\u2318 \"")
	stringEnsureGood(t, "😂 ", "\"\\u0001\\uD83D\\uDE02 \"")
	stringEnsureGood(t, `☺️`, `"\u263A\uFE0F"`)
	stringEnsureGood(t, `日本語`, `"\u65E5\u672C\u8A9E"`)

	stringEnsureBad(t, "\"\\uD83D\"", "surrogate pair", "")
	stringEnsureBad(t, "\"\\uD83D\\u\"", "surrogate pair", "u\"")
	stringEnsureBad(t, "\"\\uD83D\\uD\"", "surrogate pair", "uD\"")
	stringEnsureBad(t, "\"\\uD83D\\uDE\"", "surrogate pair", "uDE\"")
	stringEnsureBad(t, "\"\\uD83D\\uDE0\"", "invalid byte", "uDE0\"")
}

const fakeMessage = "Test logging, but use a somewhat realistic message length."

func BenchmarkNew(b *testing.B) {
	buf := make([]byte, 4096)
	for i := 0; i < b.N; i++ {
		buf = goejs.AppendEncodedJSONFromString(buf, fakeMessage)
		buf = buf[:0]
	}
}

func BenchmarkOld(b *testing.B) {
	buf := make([]byte, 4096)
	for i := 0; i < b.N; i++ {
		buf = appendEncodedJSONFromString(buf, fakeMessage)
		buf = buf[:0]
	}
}

func appendEncodedJSONFromString(buf []byte, someString string) []byte {
	buf = append(buf, '"') // prefix buffer with double quote
	for _, r := range someString {
		if escaped, ok := escapeSpecialJSON(byte(r)); ok {
			buf = append(buf, escaped...)
			continue
		}
		if r < utf8.RuneSelf && unicode.IsPrint(r) {
			buf = append(buf, byte(r))
			continue
		}
		// NOTE: Attempt to encode code point as UTF-16 surrogate pair
		r1, r2 := utf16.EncodeRune(r)
		if r1 != unicode.ReplacementChar || r2 != unicode.ReplacementChar {
			// code point does require surrogate pair, and thus two uint16 values
			buf = appendUnicodeHex(buf, uint16(r1))
			buf = appendUnicodeHex(buf, uint16(r2))
			continue
		}
		// Code Point does not require surrogate pair.
		buf = appendUnicodeHex(buf, uint16(r))
	}
	return append(buf, '"') // postfix buffer with double quote
}

// escapeSpecialJSON
func escapeSpecialJSON(b byte) ([]byte, bool) {
	// NOTE: The following 8 special JSON characters must be escaped:
	switch b {
	case '"':
		return sliceQuote, true
	case '\\':
		return sliceBackslash, true
	case '/':
		return sliceSlash, true
	case '\b':
		return sliceBackspace, true
	case '\f':
		return sliceFormfeed, true
	case '\n':
		return sliceNewline, true
	case '\r':
		return sliceCarriageReturn, true
	case '\t':
		return sliceTab, true
	}
	return nil, false
}

const hexDigits = "0123456789ABCDEF"

func appendUnicodeHex(buf []byte, v uint16) []byte {
	// Start with '\u' prefix:
	buf = append(buf, sliceUnicode...)
	// And tack on 4 hexidecimal digits:
	buf = append(buf, hexDigits[(v&0xF000)>>12])
	buf = append(buf, hexDigits[(v&0xF00)>>8])
	buf = append(buf, hexDigits[(v&0xF0)>>4])
	buf = append(buf, hexDigits[(v&0xF)])
	return buf
}

// While slices in Go are never constants, we can initialize them once and reuse
// them many times. We define these slices at library load time and reuse them
// when encoding JSON.
var (
	sliceQuote          = []byte("\\\"")
	sliceBackslash      = []byte("\\\\")
	sliceSlash          = []byte("\\/")
	sliceBackspace      = []byte("\\b")
	sliceFormfeed       = []byte("\\f")
	sliceNewline        = []byte("\\n")
	sliceCarriageReturn = []byte("\\r")
	sliceTab            = []byte("\\t")
	sliceUnicode        = []byte("\\u")
)
