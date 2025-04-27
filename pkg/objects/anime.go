package objects

import (
	"time"

	"github.com/darenliang/jikan-go"
)

type DBRecords interface {
	GetDBRecords() ([]string, []any)
	GetTblName() string
}

type Anime struct {
	jkAnime *jikan.AnimeBase
	tblName string
}

func NewAnime(data *jikan.AnimeBase) *Anime {
	return &Anime{
		jkAnime: data,
		tblName: "ANIME",
	}
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

func (a *Anime) GetDBRecords() ([]string, []any) {
	return []string{
			"ID",
			"TITLE",
			"START_DATE",
			"END_DATE",
			"STATUS",
		}, []any{
			a.GetID(),
			a.GetTitle(),
			a.GetStartDate(),
			a.GetEndDate(),
			a.GetStatus(),
		}
}

func (a *Anime) GetTblName() string {
	return a.tblName
}

type episode struct {
	AniID    int
	EpNum    int
	Duration int
	Title    string
	Aired    time.Time
	tblName  string
}

func (e *episode) GetDBRecords() ([]string, []any) {
	return []string{
			"ANI_ID",
			"EP_NUM",
			"TITLE",
			"AIR_DATE",
		}, []any{
			e.AniID,
			e.EpNum,
			e.Title,
			e.Aired,
		}
}

func (e *episode) GetTblName() string {
	return e.tblName
}

func NewEpisode(id, pId, duration int, title string, aired time.Time) *episode {
	return &episode{
		EpNum:    id,
		AniID:    pId,
		Duration: duration,
		Title:    title,
		Aired:    aired,
		tblName:  "EPISODES",
	}
}
