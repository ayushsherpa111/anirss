package api

import (
	"net/url"
	"strings"

	"github.com/darenliang/jikan-go"
)

func isMatch(query, search string) bool {
	return strings.Compare(strings.ToLower(query), strings.ToLower(search[:len(query)])) == 0
}

func GetAnimeByName(name string, anime []*jikan.AnimeBase) error {
	params := url.Values{}

	params.Set("q", name)
	params.Set("page", "1")
	params.Set("limit", "10")

	result, error := jikan.GetAnimeSearch(params)
	if error != nil {
		return error
	}

	for _, v := range result.Data {
		if isMatch(name, v.TitleEnglish) {
			anime = append(anime, &v)
		}
	}

	return nil
}
