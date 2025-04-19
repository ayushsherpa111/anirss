package utils

import (
	"encoding/xml"
	"io"

	"github.com/ayushsherpa111/anirss/pkg/objects"
)

func ParseXML(inputStream io.Reader) (objects.Anime, error) {
	var parsedBody objects.Anime
	epDecoder := xml.NewDecoder(inputStream)

	if err := epDecoder.Decode(&parsedBody); err != nil {
		return parsedBody, err
	}

	if parsedBody.EndDate.IsZero() {
		parsedBody.Status = "Ongoing"
	} else {
		parsedBody.Status = "Completed"
	}

	parsedBody.FilterEpisodes()
	return parsedBody, nil
}
