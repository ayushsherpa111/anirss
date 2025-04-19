package objects

import (
	"strings"
	"time"

	"github.com/darenliang/jikan-go"
)

type Anime struct {
	jkAnime *jikan.AnimeBase
}

type Episode struct {
	EpNum   string `json:"epno"`
	Num     int
	Title   string  `json:"title"`
	Summary string  `json:"summary"`
	AirDate xmlDate `json:"airdate"`
}

func NewAnime(data []*jikan.AnimeBase, search string) *Anime {
	var an *Anime
	for _, anime := range data {
		if strings.EqualFold(anime.TitleEnglish, search) {
			an = &Anime{
				anime,
			}
			break
		}
	}
	return an
}

func (a *Anime) GetID() int {
	return a.jkAnime.MalId
}

func (a *Anime) GetTitle() string {
	return a.jkAnime.TitleEnglish
}

func (a *Anime) GetStartDate() time.Time {
	return a.jkAnime.Aired.From
}

func (a *Anime) GetEndDate() time.Time {
	return a.jkAnime.Aired.To
}

func (a *Anime) GetStatus() string {
	return a.jkAnime.Status
}
