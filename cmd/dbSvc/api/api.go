package api

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ayushsherpa111/anirss/pkg/objects"
	"github.com/darenliang/jikan-go"
)

func isMatch(query, search string) bool {
	return strings.Compare(strings.ToLower(query), strings.ToLower(search)) == 0
}

func GetAnimeByName(name string) (*objects.Anime, error) {
	params := url.Values{}

	params.Set("q", name)
	params.Set("page", "1")
	params.Set("limit", "10")

	result, error := jikan.GetAnimeSearch(params)
	if error != nil {
		return nil, error
	}

	for _, v := range result.Data {
		if isMatch(name, v.TitleEnglish) {
			// prevent rate limiting
			time.Sleep(time.Second)
			return objects.NewAnime(&v), nil
		}
	}

	return nil, fmt.Errorf("could not find any anime of the name %s", name)
}

func GetAnimeEpisodes(wg *sync.WaitGroup, id int, page int, filteredEpisodes chan objects.DBRecords) error {
	wg.Add(1)
	episodes, err := jikan.GetAnimeEpisodes(id, page)
	// prevent rate limiting
	time.Sleep(time.Second)
	if err != nil {
		return err
	}
	go func() {
		for _, ep := range episodes.Data {
			if !ep.Filler || !ep.Recap {
				filteredEpisodes <- &objects.Episode{
					AniID:    id,
					ID:       ep.MalId,
					Duration: ep.Duration,
					Title:    ep.Title,
					Aired:    ep.Aired,
				}
			}
		}
		wg.Done()
	}()

	if episodes.Pagination.HasNextPage {
		err = GetAnimeEpisodes(wg, id, page+1, filteredEpisodes)
	}
	return err
}
