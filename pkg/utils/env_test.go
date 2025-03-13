package utils

import (
	"bytes"
	"os"
	"regexp"
	"testing"
)

func TestParseVal(t *testing.T) {
	os.Setenv("Hello", "hi")
	inputVal := [][]byte{
		[]byte("This is $Hello"),
		[]byte(`"This is $Hello"`),
		[]byte("'This is $Hello'"),
	}
	expect := [][]byte{
		[]byte("This is hi"),
		[]byte("This is hi"),
		[]byte("This is $Hello"),
	}
	quoteRgx := regexp.MustCompile(quotRegexPat)
	varRgx := regexp.MustCompile(varRegexPat)

	for i, v := range inputVal {
		if got, err := parseVal(quoteRgx, varRgx, v); err != nil {
			t.Errorf("Error: %s. Parsing %s", err.Error(), v)
		} else if !bytes.Equal([]byte(got), expect[i]) {
			t.Errorf("Invalid value parsed. Expected %q Got %q", expect[i], got)
		}
	}
}
