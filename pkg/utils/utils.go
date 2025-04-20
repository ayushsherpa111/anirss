package utils

import (
	"fmt"
	"net/url"
	"sync"
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

func Multiplexer[T any](inpChan ...chan T) chan T {
	multiplexChan := make(chan T)
	wg := sync.WaitGroup{}
	wg.Add(len(inpChan))
	multiplexFunc := func(c chan T) {
		defer wg.Done()
		for v := range c {
			multiplexChan <- v
		}
	}

	for _, c := range inpChan {
		go multiplexFunc(c)
	}

	go func() {
		wg.Wait()
		close(multiplexChan)
	}()

	return multiplexChan
}
