package goejs_test

import (
	"strings"
	"testing"
)

func ensureError(tb testing.TB, testCase string, err error, contains string) {
	if err == nil || !strings.Contains(err.Error(), contains) {
		tb.Errorf("Case: %#q; Actual: %v; Expected: %s", testCase, err, contains)
	}
}
