package goejs

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
)

// AppendEncodedJSONFromFloat appends the JSON encoded form of the value to the
// provided byte slice. Because some legal IEEE 754 floating point values have
// no JSON equivalents, this library encodes several floating point numbers into
// the corresponding encoded form, as used by several other JSON encoding
// libraries, as shown in the table below.
//
// JSON serialization:
//    NaN: null
//   -Inf: -1e999
//   +Inf: 1e999
func AppendEncodedJSONFromFloat(buf []byte, f64 float64) []byte {
	if math.IsNaN(f64) {
		return append(buf, "null"...)
	} else if math.IsInf(f64, 1) {
		return append(buf, "1e999"...)
	} else if math.IsInf(f64, -1) {
		return append(buf, "-1e999"...)
	}
	// NOTE: To support some dynamic languages which will decode a JSON number
	// without a fractional component as a runtime integer, we encode these
	// numbers using exponential notation.
	if f64 == math.Floor(f64) {
		return strconv.AppendFloat(buf, f64, 'e', -1, 64)
	}
	// Otherwise, use the most compact format possible.
	return strconv.AppendFloat(buf, f64, 'g', -1, 64)
}

// DecodedFloatFromJSON consumes bytes from the provided byte slice to decode a
// floating point number. Because some legal IEEE 754 floating point values have
// no JSON equivalents, this library decodes the following encoded forms, as
// used by several other JSON encoding libraries, into floating point numbers as
// shown in the table below.
//
// JSON serialization:
//    NaN: null
//   -Inf: -1e999
//   +Inf: 1e999
func DecodedFloatFromJSON(buf []byte) (float64, []byte, error) {
	buflen := len(buf)
	if buflen >= 4 {
		if bytes.Equal(buf[:4], []byte("null")) {
			return math.NaN(), buf[4:], nil
		}
		if buflen >= 5 {
			if bytes.Equal(buf[:5], []byte("1e999")) {
				return math.Inf(1), buf[5:], nil
			}
			if buflen >= 6 {
				if bytes.Equal(buf[:6], []byte("-1e999")) {
					return math.Inf(-1), buf[6:], nil
				}
			}
		}
	}
	index, err := numberLength(buf, true) // NOTE: floatAllowed = true
	if err != nil {
		return 0, buf, err
	}
	datum, err := strconv.ParseFloat(string(buf[:index]), 64)
	if err != nil {
		return 0, buf, err
	}
	return datum, buf[index:], nil
}

func numberLength(buf []byte, floatAllowed bool) (int, error) {
	// ALGORITHM: increment index as long as bytes are valid for number state engine.
	var index, buflen, count int
	var b byte

	// STATE 0: begin, optional: -
	if buflen = len(buf); index == buflen {
		return 0, io.ErrShortBuffer
	}
	if buf[index] == '-' {
		if index++; index == buflen {
			return 0, io.ErrShortBuffer
		}
	}
	// STATE 1: if 0, goto 2; otherwise if 1-9, goto 3; otherwise bail
	if b = buf[index]; b == '0' {
		if index++; index == buflen {
			return index, nil // valid number
		}
	} else if b >= '1' && b <= '9' {
		if index++; index == buflen {
			return index, nil // valid number
		}
		// STATE 3: absorb zero or more digits
		for {
			if b = buf[index]; b < '0' || b > '9' {
				break
			}
			if index++; index == buflen {
				return index, nil // valid number
			}
		}
	} else {
		return 0, fmt.Errorf("unexpected byte: %q", b)
	}
	if floatAllowed {
		// STATE 2: if ., goto 4; otherwise goto 5
		if buf[index] == '.' {
			if index++; index == buflen {
				return 0, io.ErrShortBuffer
			}
			// STATE 4: absorb one or more digits
			for {
				if b = buf[index]; b < '0' || b > '9' {
					break
				}
				count++
				if index++; index == buflen {
					return index, nil // valid number
				}
			}
			if count == 0 {
				// did not get at least one digit
				return 0, fmt.Errorf("unexpected byte: %q", b)
			}
		}
		// STATE 5: if e|e, goto 6; otherwise goto 7
		if b = buf[index]; b == 'e' || b == 'E' {
			if index++; index == buflen {
				return 0, io.ErrShortBuffer
			}
			// STATE 6: if -|+, goto 8; otherwise goto 8
			if b = buf[index]; b == '+' || b == '-' {
				if index++; index == buflen {
					return 0, io.ErrShortBuffer
				}
			}
			// STATE 8: absorb one or more digits
			count = 0
			for {
				if b = buf[index]; b < '0' || b > '9' {
					break
				}
				count++
				if index++; index == buflen {
					return index, nil // valid number
				}
			}
			if count == 0 {
				// did not get at least one digit
				return 0, fmt.Errorf("unexpected byte: %q", b)
			}
		}
	}
	// STATE 7: end
	return index, nil
}

// DecodedIntFromJSON consumes bytes from the provided byte slice to decode an
// integer number.
func DecodedIntFromJSON(buf []byte) (int64, []byte, error) {
	index, err := numberLength(buf, false) // NOTE: floatAllowed = false
	if err != nil {
		return 0, buf, err
	}
	datum, err := strconv.ParseInt(string(buf[:index]), 10, 64)
	if err != nil {
		return 0, buf, err
	}
	return datum, buf[index:], nil
}

// AppendEncodedJSONFromInt appends the JSON encoded form of the value to the
// provided byte slice. While this function merely wraps a standard library
// function, it is provided for API symmetry.
func AppendEncodedJSONFromInt(buf []byte, i int64) []byte {
	return strconv.AppendInt(buf, i, 10)
}
