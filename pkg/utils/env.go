package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	keyRegexPat  = "[a-zA-Z_]+[a-zA-Z0-9_]*"              // Regex for Keys of .env files
	varRegexPat  = fmt.Sprintf(`\${?(%s)}?`, keyRegexPat) // Regex for Variable To Interpolate
	quotRegexPat = `(['"])?([^'"]*)(['"])?`
)

func LoadEnv(lgr *slog.Logger, fileName ...string) {
	// use logger
	if len(fileName) == 0 {
		// look for .env in root
		fileName = append(fileName, ".env")
	}

	dir, err := os.Getwd()
	if err != nil {
		lgr.Error("Failed to get cwd.", "ErrMsg", err.Error())
		// log error
		os.Exit(1)
	}

	for _, file := range fileName {
		envPath := path.Join(dir, file)
		if err := parseEnvFile(envPath, lgr); err != nil {
			// log error explaining that the file either doesnot exist or is a directory
			lgr.Error("error encountred while parsing ENV file.", "ErrMsg", err.Error())
			os.Exit(1)
		}
	}
}

func parseVal(quoteRgx *regexp.Regexp, varRgx *regexp.Regexp, val []byte) (string, error) {
	results := quoteRgx.FindAllSubmatch(val, -1)[0]
	oQuote, value, cQuote := results[1], results[2], results[3]
	sValue := string(value)

	if !bytes.Equal(oQuote, cQuote) {
		return "", fmt.Errorf("mismatched quote. Expected (%s and %s), got (%s and %s)", oQuote, oQuote, oQuote, cQuote)
	}

	// Interpolate any variables found in the quote or no quotes
	if bytes.Equal(oQuote, []byte("\"")) || len(oQuote) == 0 {
		result := varRgx.FindAllSubmatch(value, -1)
		for _, keySlc := range result {
			key := keySlc[1]
			repl := os.Getenv(string(key))
			sValue = strings.ReplaceAll(sValue, string(keySlc[0]), repl)
		}
	}

	return strings.Trim(sValue, "\n"), nil
}

func isLineValid(line []byte) bool {
	// check if  the line is a comment
	comment := byte('#')
	re := regexp.MustCompile(`\\s`)
	sanitized := re.ReplaceAll(line, []byte{})

	// is a blank line or only contains whitespace
	if len(sanitized) == 0 || sanitized[0] == comment {
		return false
	}

	return true
}

func parseEnvFile(path string, lgr *slog.Logger) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	keyRgx := regexp.MustCompile(keyRegexPat)
	quoteRgx := regexp.MustCompile(quotRegexPat)
	varRgx := regexp.MustCompile(varRegexPat)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			// handle err
			lgr.Error("Error reading from file", "readErr", err.Error())
			break
		}

		if !isLineValid(line) {
			continue
		}

		// split line by = sign.
		// set the ENV variable using the key
		envInp := bytes.Split(line, []byte("="))
		if match := keyRgx.Match(envInp[0]); match {
			if envVal, err := parseVal(quoteRgx, varRgx, envInp[1]); err != nil {
				// log error
				lgr.Error("error parsing .env", "key", string(envInp[0]), "val", envVal)
			} else {
				lgr.Info("setting Env", "key", string(envInp[0]), "value", envVal)
				os.Setenv(string(envInp[0]), envVal)
			}
		} else {
			// log a skipped ENV
			lgr.Info("skipping setting env.", "line", line)
		}
	}
	return nil
}
