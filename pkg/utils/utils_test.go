package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestAddParams(t *testing.T) {
	aniURI := "http://api.anidb.net:9001/httpapi"
	inputVal := [][]string{
		{"aid", "1"},
		{"anime", "solo", "client", "1", "ver", "2"},
	}
	expect := []string{
		fmt.Sprintf("%s?aid=1", aniURI),
		fmt.Sprintf("%s?anime=solo&client=1&ver=2", aniURI),
	}
	for i, v := range inputVal {
		if result, err := AddParams(aniURI, v...); err != nil {
			t.Errorf("error adding params %s\n", err)
		} else if strings.Compare(expect[i], result) != 0 {
			t.Errorf("values do not match. expected: %s, got: %s", expect[i], result)
		} else {
			fmt.Printf("passed.\nExpected: %s, Got: %s\n", expect[i], result)
		}
	}
}
