package goejs

import "unicode/utf8"

// AppendEncodedJSONFromString appends the JSON encoding of the provided string to the
// provided byte slice, and returns the modified byte slice.
//
//    func ExampleEncode() {
//        encoded := AppendEncodedJSONFromString([]byte("prefix:"), "⌘ a")
//        fmt.Printf("%s", encoded)
//        // Output: prefix:"\u0001\u2318 a"
//    }
func AppendEncodedJSONFromString(buf []byte, someString string) []byte {
	buf = append(buf, '"') // prefix buffer with double quote

	for _, r := range someString {
		if r < utf8.RuneSelf {
			i8 := special[byte(r)]
			if i8 < 0 {
				buf = append(buf, '\\')
				buf = append(buf, uint8(-i8))
				continue
			}
			if i8 > 0 {
				buf = append(buf, uint8(i8))
				continue
			}
		}

		if r < surrSelf || r > maxRune {
			// This rune is encoded using "\uXXXX" notation.
			u16 := uint16(r)
			buf = append(buf, sliceUnicode...)
			buf = append(buf, hexDigits[(u16&0xF000)>>12])
			buf = append(buf, hexDigits[(u16&0xF00)>>8])
			buf = append(buf, hexDigits[(u16&0xF0)>>4])
			buf = append(buf, hexDigits[(u16&0xF)])
			continue
		}

		// This rune requires encoding using a surrogate pair of code points.
		r -= surrSelf

		u1 := uint16(surr1 + (r>>10)&0x3ff)
		buf = append(buf, sliceUnicode...)
		buf = append(buf, hexDigits[(u1&0xF000)>>12])
		buf = append(buf, hexDigits[(u1&0xF00)>>8])
		buf = append(buf, hexDigits[(u1&0xF0)>>4])
		buf = append(buf, hexDigits[(u1&0xF)])

		u2 := uint16(surr2 + r&0x3ff)
		buf = append(buf, sliceUnicode...)
		buf = append(buf, hexDigits[(u2&0xF000)>>12])
		buf = append(buf, hexDigits[(u2&0xF00)>>8])
		buf = append(buf, hexDigits[(u2&0xF0)>>4])
		buf = append(buf, hexDigits[(u2&0xF)])
	}

	return append(buf, '"') // postfix buffer with double quote
}

const (
	hexDigits       = "0123456789ABCDEF"
	replacementChar = '\uFFFD'     // Unicode replacement character
	maxRune         = '\U0010FFFF' // Maximum valid Unicode code point.
)

const (
	// 0xd800-0xdc00 encodes the high 10 bits of a pair.
	// 0xdc00-0xe000 encodes the low 10 bits of a pair.
	// the value is those 20 bits plus 0x10000.
	surr1 = 0xd800
	surr2 = 0xdc00
	surr3 = 0xe000

	surrSelf = 0x10000
)

// // While slices in Go are never constants, we can initialize them once and
// // reuse them many times. We define these slices at library load time and
// // reuse them when encoding JSON.
// var sliceUnicode = []byte("\\u")

// special is a list of values to encode each rune. When the value is
// positive, then the single byte should be emitted to the stream. When the
// value is negative, then the negative of the value should be emitted to the
// stream after a prefix backslash character. When the value is 0, then value
// should be encoded using "\uXXXX" notation.
var special = [utf8.RuneSelf + 1]int8{
	0x00: 0,    // '\x00',
	0x01: 0,    // '\x01',
	0x02: 0,    // '\x02',
	0x03: 0,    // '\x03',
	0x04: 0,    // '\x04',
	0x05: 0,    // '\x05',
	0x06: 0,    // '\x06',
	0x07: 0,    // '\a',
	0x08: -'b', // '\b',
	0x09: -'t', // '\t',
	0x0A: -'n', // '\n',
	0x0B: '\v',
	0x0C: -'f', // '\f',
	0x0D: -'r', // '\r',
	0x0E: 0,    // '\x0e',
	0x0F: 0,    // '\x0f',
	0x10: 0,    // '\x10',
	0x11: 0,    // '\x11',
	0x12: 0,    // '\x12',
	0x13: 0,    // '\x13',
	0x14: 0,    // '\x14',
	0x15: 0,    // '\x15',
	0x16: 0,    // '\x16',
	0x17: 0,    // '\x17',
	0x18: 0,    // '\x18',
	0x19: 0,    // '\x19',
	0x1A: 0,    // '\x1a',
	0x1B: 0,    // '\x1b',
	0x1C: 0,    // '\x1c',
	0x1D: 0,    // '\x1d',
	0x1E: 0,    // '\x1e',
	0x1F: 0,    // '\x1f',
	0x20: ' ',
	0x21: '!',
	0x22: -'"', // '"',
	0x23: '#',
	0x24: '$',
	0x25: '%',
	0x26: '&',
	0x27: '\'',
	0x28: '(',
	0x29: ')',
	0x2A: '*',
	0x2B: '+',
	0x2C: ',',
	0x2D: '-',
	0x2E: '.',
	0x2F: -'/', // '/',
	0x30: '0',
	0x31: '1',
	0x32: '2',
	0x33: '3',
	0x34: '4',
	0x35: '5',
	0x36: '6',
	0x37: '7',
	0x38: '8',
	0x39: '9',
	0x3A: ':',
	0x3B: ';',
	0x3C: '<',
	0x3D: '=',
	0x3E: '>',
	0x3F: '?',
	0x40: '@',
	0x41: 'A',
	0x42: 'B',
	0x43: 'C',
	0x44: 'D',
	0x45: 'E',
	0x46: 'F',
	0x47: 'G',
	0x48: 'H',
	0x49: 'I',
	0x4A: 'J',
	0x4B: 'K',
	0x4C: 'L',
	0x4D: 'M',
	0x4E: 'N',
	0x4F: 'O',
	0x50: 'P',
	0x51: 'Q',
	0x52: 'R',
	0x53: 'S',
	0x54: 'T',
	0x55: 'U',
	0x56: 'V',
	0x57: 'W',
	0x58: 'X',
	0x59: 'Y',
	0x5A: 'Z',
	0x5B: '[',
	0x5C: -'\\', // '\\',
	0x5D: ']',
	0x5E: '^',
	0x5F: '_',
	0x60: '`',
	0x61: 'a',
	0x62: 'b',
	0x63: 'c',
	0x64: 'd',
	0x65: 'e',
	0x66: 'f',
	0x67: 'g',
	0x68: 'h',
	0x69: 'i',
	0x6A: 'j',
	0x6B: 'k',
	0x6C: 'l',
	0x6D: 'm',
	0x6E: 'n',
	0x6F: 'o',
	0x70: 'p',
	0x71: 'q',
	0x72: 'r',
	0x73: 's',
	0x74: 't',
	0x75: 'u',
	0x76: 'v',
	0x77: 'w',
	0x78: 'x',
	0x79: 'y',
	0x7A: 'z',
	0x7B: '{',
	0x7C: '|',
	0x7D: '}',
	0x7E: '~',
}
