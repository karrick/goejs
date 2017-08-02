package goejs_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/karrick/goejs"
)

func floatEnsureBad(tb testing.TB, input, errorMessage, remainder string) {
	_, buf2, err := goejs.DecodedFloatFromJSON([]byte(input))
	ensureError(tb, input, err, errorMessage)
	if actual, expected := string(buf2), remainder; actual != expected {
		tb.Errorf("Input: %#q; Remainder Actual: %#q; Expected: %#q; Error: %s", input, actual, expected, err)
	}
}

func floatEnsureGood(tb testing.TB, input float64, expected string) {
	buf := goejs.AppendEncodedJSONFromFloat(nil, input)
	if actual := string(buf); actual != expected {
		tb.Errorf("Input: %#q; Actual: %#q; Expected: %#q", input, actual, expected)
	}
	output, buf2, err := goejs.DecodedFloatFromJSON([]byte(expected))
	if err != nil {
		tb.Errorf("Input: %#q: %s", input, err)
	}
	if math.IsNaN(input) != math.IsNaN(output) && math.IsInf(input, 1) != math.IsInf(output, 1) && math.IsInf(input, -1) != math.IsInf(output, -1) && input != output {
		tb.Errorf("Input: %#q; Output: %#q", input, output)
	}
	if actual, expected := string(buf2), ""; actual != expected {
		tb.Errorf("Input: %#q; Remainder Actual: %#q; Expected: %#q", input, actual, expected)
	}
}

func ExampleFloatDecode() {
	decoded, remainder, err := goejs.DecodedFloatFromJSON([]byte("null some extra bytes after final quote"))
	if err != nil {
		fmt.Println(err)
	}
	if actual, expected := string(remainder), " some extra bytes after final quote"; actual != expected {
		fmt.Printf("Remainder Actual: %#q; Expected: %#q\n", actual, expected)
	}
	fmt.Printf("%v", decoded)
	// Output: NaN
}

func ExampleFloatEncode() {
	encoded := goejs.AppendEncodedJSONFromFloat([]byte("prefix: "), math.NaN())
	fmt.Printf("%s", encoded)
	// Output: prefix: null
}

func TestFloat(t *testing.T) {
	floatEnsureBad(t, "", "short buffer", "")
	floatEnsureBad(t, "foo", "unexpected byte: 'f'", "foo")

	floatEnsureGood(t, 10, "1e+01") // whole numbers represented in `e` notation
	floatEnsureGood(t, 3.14, "3.14")
	floatEnsureGood(t, math.NaN(), "null")
	floatEnsureGood(t, math.Inf(1), "1e999")
	floatEnsureGood(t, math.Inf(-1), "-1e999")
}
