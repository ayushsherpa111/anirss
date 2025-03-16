package dbhandler

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/ayushsherpa111/anirss/pkg/rpc/dbservice"
	"github.com/ayushsherpa111/anirss/pkg/utils"
)

const (
	aniLookup = "http://api.anidb.net:9001/httpapi" // ani Lookeup for list of episodes
)

type dbRPC struct {
	dbservice.UnimplementedAniDbSvcServer // fallback for unimpemented RPC functions
	db                                    *sql.DB
	logger                                *slog.Logger
	httpClient                            *http.Client
	dbLookupQuery                         string
}

func (db *dbRPC) getAnime(ctx context.Context, aniId int) {
	parsedURI, err := url.Parse(db.dbLookupQuery)
	if err != nil {
		db.logger.Error("failed to parse URI", "err", err.Error())
		panic(err)
	}
	rawURI := parsedURI.Query()
	rawURI.Add("aid", strconv.Itoa(aniId))
	parsedURI.RawQuery = rawURI.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", parsedURI.String(), nil)
	if err != nil {
		db.logger.Error("failed to create new request", "err", err.Error())
	}
	resp, err := db.httpClient.Do(req)
	if err != nil {
		db.logger.Error("received response was invalid.", "err", err.Error())
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		db.logger.Error("Response code not OK", "err", resp.StatusCode)
		panic(fmt.Errorf("received a non 200 status code. Status Code: %d", resp.StatusCode))
	}
	utils.ParseXML(resp.Body)
}

func (d *dbRPC) AddAnime(ctx context.Context, anime *dbservice.AniParams) (*dbservice.Result, error) {
	result := new(dbservice.Result)
	// HTTP request
	d.getAnime(ctx, int(anime.GetAnimeID()))
	return result, nil
}

func NewDBServer(dbLogger *slog.Logger) *dbRPC {
	dbRpcSvc := new(dbRPC)
	dbRpcSvc.db = InitDB(dbLogger)
	dbRpcSvc.logger = dbLogger
	dbRpcSvc.httpClient = &http.Client{}

	anidbURI, err := url.Parse(aniLookup)
	if err != nil {
		dbLogger.Error("failed to parse lookup URL", "err", err.Error())
		os.Exit(1)
	}

	client := os.Getenv("ANI_CLIENT")
	clientver := os.Getenv("ANI_CLIENTVER")
	protocol := os.Getenv("ANI_PROTOVER")
	request := os.Getenv("ANI_REQUEST")

	queryParams := anidbURI.Query()
	// add required params to make the http request
	queryParams.Add("client", client)
	queryParams.Add("clientver", clientver)
	queryParams.Add("protover", protocol)
	queryParams.Add("request", request)

	anidbURI.RawQuery = queryParams.Encode()

	dbRpcSvc.dbLookupQuery = anidbURI.String()
	return dbRpcSvc
}
