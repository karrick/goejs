package goejs

// func parseFloat(buf []byte, floatAllowed bool) (float64, error) {
// 	// ALGORITHM: increment index as long as bytes are valid for number state engine.
// 	var index, buflen, count int
// 	var b byte

// 	var mantissa int
// 	var isMantissaNegative bool
// 	var fractional int
// 	var isExponentNegative bool
// 	var exponent int

// 	// STATE 0: begin, optional: -
// 	if buflen = len(buf); index == buflen {
// 		return 0, io.ErrShortBuffer
// 	}
// 	if buf[index] == '-' {
// 		if index++; index == buflen {
// 			return 0, io.ErrShortBuffer
// 		}
// 		isMantissaNegative = true
// 	}
// 	// STATE 1: if 0, goto 2; otherwise if 1-9, goto 3; otherwise bail
// 	if b = buf[index]; b == '0' {
// 		if index++; index == buflen {
// 			return 0, nil // valid number
// 		}
// 	} else if b >= '1' && b <= '9' {
// 		mantissa = int(b - '0')
// 		if index++; index == buflen {
// 			return float64(mantissa), nil
// 		}
// 		// STATE 3: absorb zero or more digits
// 		for {
// 			if b = buf[index]; b < '0' || b > '9' {
// 				break
// 			}
// 			mantissa = 10*mantissa + int(b-'0')
// 			if index++; index == buflen {
// 				return float64(mantissa), nil // valid number
// 			}
// 		}
// 	} else {
// 		return 0, fmt.Errorf("unexpected byte: %#U", b)
// 	}
// 	if floatAllowed {
// 		// STATE 2: if ., goto 4; otherwise goto 5
// 		if buf[index] == '.' {
// 			if index++; index == buflen {
// 				return 0, io.ErrShortBuffer
// 			}
// 			// STATE 4: absorb one or more digits
// 			for {
// 				if b = buf[index]; b < '0' || b > '9' {
// 					break
// 				}
// 				fractional = 10*fractional + int(b-'0')
// 				count++
// 				if index++; index == buflen {
// 					return index, nil // valid number
// 				}
// 			}
// 			if count == 0 {
// 				// did not get at least one digit
// 				return 0, fmt.Errorf("unexpected byte: %q", b)
// 			}
// 		}
// 		// STATE 5: if e|e, goto 6; otherwise goto 7
// 		if b = buf[index]; b == 'e' || b == 'E' {
// 			if index++; index == buflen {
// 				return 0, io.ErrShortBuffer
// 			}
// 			// STATE 6: if -|+, goto 8; otherwise goto 8
// 			if b = buf[index]; b == '+' || b == '-' {
// 				if index++; index == buflen {
// 					return 0, io.ErrShortBuffer
// 				}
// 			}
// 			// STATE 8: absorb one or more digits
// 			count = 0
// 			for {
// 				if b = buf[index]; b < '0' || b > '9' {
// 					break
// 				}
// 				count++
// 				if index++; index == buflen {
// 					return index, nil // valid number
// 				}
// 			}
// 			if count == 0 {
// 				// did not get at least one digit
// 				return 0, fmt.Errorf("unexpected byte: %q", b)
// 			}
// 		}
// 	}
// 	// STATE 7: end
// 	return index, nil
// }
