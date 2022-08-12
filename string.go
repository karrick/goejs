package goejs

// NOTE: The contents of this file are modified versions of files from
// https://github.com/karrick/goavro/bytes.go

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"unicode/utf16"
	"unicode/utf8"
)

// While slices in Go are never constants, we can initialize them once and reuse
// them many times. We define these slices at library load time and reuse them
// when encoding JSON.
var (
	sliceUnicode = []byte("\\u")
)

// DecodedStringFromJSON decodes a string from JSON, returning the decoded
// string and the remainder byte slice of the original buffer. On error, the
// returned byte slice points to the first byte that caused the error indicated.
//
//    func ExampleDecode() {
//        decoded, remainder, err := goejs.DecodedStringFromJSON([]byte("\"\\u0001\\u2318 a\" some extra bytes after final quote"))
//        if err != nil {
//            fmt.Println(err)
//        }
//        if actual, expected := string(remainder), " some extra bytes after final quote"; actual != expected {
//            fmt.Printf("Remainder Actual: %#q; Expected: %#q\n", actual, expected)
//        }
//        fmt.Printf("%#q", decoded)
//        // Output: "\x01⌘ a"
//    }
func DecodedStringFromJSON(buf []byte) (string, []byte, error) {
	buflen := len(buf)
	if buflen < 2 {
		return "", buf, fmt.Errorf("cannot decode string: %s", io.ErrShortBuffer)
	}
	if buf[0] != '"' {
		return "", buf, fmt.Errorf("cannot decode string: expected initial '\"'; found: %#U", buf[0])
	}
	var newBytes []byte
	var escaped, ok bool
	// Loop through bytes following initial double quote, but note we will
	// return immediately when find unescaped double quote.
	for i := 1; i < buflen; i++ {
		b := buf[i]
		if escaped {
			escaped = false
			if b, ok = unescapeSpecialJSON(b); ok {
				newBytes = append(newBytes, b)
				continue
			}
			if b == 'u' {
				// NOTE: Need at least 4 more bytes to read uint16, but subtract
				// 1 because do not want to count the trailing quote and
				// subtract another 1 because already consumed u but have yet to
				// increment i.
				if i > buflen-6 {
					return "", buf[i+1:], fmt.Errorf("cannot decode string: %s", io.ErrShortBuffer)
				}
				v, err := parseUint64FromHexSlice(buf[i+1 : i+5])
				if err != nil {
					return "", buf[i+1:], fmt.Errorf("cannot decode string: %s", err)
				}
				i += 4 // absorb 4 characters: one 'u' and three of the digits

				nbl := len(newBytes)
				newBytes = append(newBytes, 0, 0, 0, 0) // grow to make room for UTF-8 encoded rune

				r := rune(v)
				if utf16.IsSurrogate(r) {
					i++ // absorb final hexidecimal digit from previous value

					// Expect second half of surrogate pair
					if i > buflen-6 || buf[i] != '\\' || buf[i+1] != 'u' {
						return "", buf[i+1:], errors.New("cannot decode string: missing second half of surrogate pair")
					}

					v, err = parseUint64FromHexSlice(buf[i+2 : i+6])
					if err != nil {
						return "", buf[i+1:], fmt.Errorf("cannot decode string: cannot decode second half of surrogate pair: %s", err)
					}
					i += 5 // absorb 5 characters: two for '\u', and 3 of the 4 digits

					// Get code point by combining high and low surrogate bits
					r = utf16.DecodeRune(r, rune(v))
				}

				width := utf8.EncodeRune(newBytes[nbl:], r) // append UTF-8 encoded version of code point
				newBytes = newBytes[:nbl+width]             // trim off excess bytes
				continue
			}
			newBytes = append(newBytes, b)
			continue
		}
		if b == '\\' {
			escaped = true
			continue
		}
		if b == '"' {
			return string(newBytes), buf[i+1:], nil
		}
		newBytes = append(newBytes, b)
	}
	return "", buf, fmt.Errorf("cannot decode string: expected final '\"'; found: %#U", buf[buflen-1])
}

// parseUint64FromHexSlice decodes four characters as hexidecimal digits into a
// uint64 value. It returns an error when any of the four characters are not
// valid hexidecimal digits.
func parseUint64FromHexSlice(buf []byte) (uint64, error) {
	var value uint64
	for _, b := range buf {
		diff := uint64(b - '0')
		if diff < 0 {
			return 0, hex.InvalidByteError(b)
		}
		if diff < 10 {
			// digit 0-9
			value = (value << 4) | diff
			continue
		}
		// letter a-f or A-F
		b10 := b + 10
		diff = uint64(b10 - 'A')
		if diff < 10 {
			return 0, hex.InvalidByteError(b)
		}
		if diff < 16 {
			// letter A-F
			value = (value << 4) | diff
			continue
		}
		// letter a-f
		diff = uint64(b10 - 'a')
		if diff < 10 {
			return 0, hex.InvalidByteError(b)
		}
		if diff < 16 {
			value = (value << 4) | diff
			continue
		}
		return 0, hex.InvalidByteError(b)
	}
	return value, nil
}

// unescapeSpecialJSON attempts to decode one of 8 special bytes. It returns the
// decoded byte and true if the original byte was one of the 8; otherwise it
// returns the original byte and false.
func unescapeSpecialJSON(b byte) (byte, bool) {
	// NOTE: The following 8 special JSON characters must be escaped:
	switch b {
	case '"', '\\', '/':
		return b, true
	case 'b':
		return '\b', true
	case 'f':
		return '\f', true
	case 'n':
		return '\n', true
	case 'r':
		return '\r', true
	case 't':
		return '\t', true
	}
	return b, false
}
