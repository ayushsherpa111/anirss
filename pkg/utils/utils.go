package utils

import (
	"fmt"
	"net/url"
)

func AddParams(uri string, args ...string) (string, error) {
	if argLen := len(args); argLen%2 != 0 {
		return "", fmt.Errorf("expected number of key value pair in args to be equal. got : %d number of arguments", argLen)
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	raw := parsed.Query()
	for i := 0; i < len(args); i += 2 {
		raw.Add(args[i], args[i+1])
	}

	parsed.RawQuery = raw.Encode()
	return parsed.String(), nil
}

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
