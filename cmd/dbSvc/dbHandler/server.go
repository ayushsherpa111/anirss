package dbhandler

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ayushsherpa111/anirss/cmd/dbSvc/api"
	"github.com/ayushsherpa111/anirss/pkg/objects"
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
	maxWorkers                            int
}

func (db *dbRPC) getAnime(aniId int) (body io.Reader) {
	context, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	queryWithID := utils.Must(utils.AddParams(db.dbLookupQuery, "aid", strconv.Itoa(aniId)))
	db.logger.Info("Lookup URI", "uri", db.dbLookupQuery, "parsed", queryWithID)
	req, err := http.NewRequestWithContext(context, http.MethodGet, queryWithID, nil)
	if err != nil {
		db.logger.Info("Failed to create new request", "err", err.Error())
	}
	db.logger.Info("Request prepared")
	// maybe try to retry http requests?
	resp, err := db.httpClient.Do(req)
	if err != nil {
		db.logger.Info("failed to create http request", "err", err)
	}

	if resp.StatusCode != http.StatusOK {
		db.logger.Error("Response code not OK", "err", resp.StatusCode)
		return
	}
	body, writer := io.Pipe()
	go func() {
		defer resp.Body.Close()
		defer writer.Close()
		bWritten, err := io.Copy(writer, resp.Body)
		db.logger.Info("Written bytes", "bytes", bWritten)
		if err != nil {
			panic(err)
		}
	}()

	return
}

func (db *dbRPC) httpStage(_ <-chan bool, idChan <-chan int) <-chan io.Reader {
	bodyChan := make(chan io.Reader)
	go func() {
		defer close(bodyChan)
		defer func() {
			if r := recover(); r != nil {
				db.logger.Debug("Recovered error in http stage", "err", r)
			}
		}()
		for id := range idChan {
			db.logger.Info("Sending request", "id", id)
			bodyChan <- db.getAnime(id)
		}
	}()
	return bodyChan
}

func (db *dbRPC) idStage(done <-chan bool, aniIds ...int) <-chan int {
	outbound := make(chan int)
	go func() {
		defer close(outbound)
		for _, aniId := range aniIds {
			select {
			case <-done:
				return
			case outbound <- aniId:
			}
		}
	}()
	return outbound
}

func (d *dbRPC) AddAnimeById(ctx context.Context, anime *dbservice.AniParams) (*dbservice.Result, error) {
	result := new(dbservice.Result)
	done := make(chan bool)
	defer close(done)
	defer func() {
		if r := recover(); r != nil {
			d.logger.Debug("recovered an error in AddAnime", "panic", r)
			result.Status = "Failed"
			result.NewEntries = 0
		}
	}()

	// d.logger.Info("Adding anime", "args", int(anime.AnimeID))
	// idChan := d.idStage(done, int(anime.AnimeID))
	// httpChan := d.httpStage(done, idChan)
	// // xmlChan := d.xmlStage(done, httpChan)
	//
	// // result.NewEntries = int32(<-d.dbStage(done, xmlChan))
	// result.Status = "Added"

	return result, nil
}

func (d *dbRPC) loggingStage() chan objects.Logging {
	loggingChan := make(chan objects.Logging, 10)
	go func() {
		for log := range loggingChan {
			fmt.Println(log.Message)
			fmt.Println(log.Payload)
			switch log.Level {
			case objects.L_ERROR:
				d.logger.Error(log.Message, "error", log.Error.Error())
			case objects.L_INFO:
				d.logger.Info(log.Message, "params", log.Payload)
			}
		}
	}()
	return loggingChan
}

func (d *dbRPC) episodeStage(aniID int) chan objects.DBRecords {
	epChan := make(chan objects.DBRecords)
	go func() {
		wg := &sync.WaitGroup{}
		defer close(epChan)
		if e := api.GetAnimeEpisodes(wg, aniID, 1, epChan); e != nil {
			d.logger.Error("Failed to fetch all episode", "error", e.Error())
		}
		wg.Wait()
	}()
	return epChan
}

func (d *dbRPC) AddAnimeByName(ctx context.Context, param *dbservice.AniParams) (*dbservice.Result, error) {
	result := new(dbservice.Result)
	loggingChan := d.loggingStage()
	doneChan := make(chan bool)
	defer close(loggingChan)
	defer close(doneChan)

	defer func() {
		if r := recover(); r != nil {
			d.logger.Debug("recovered an error in AddAnime", "panic", r)
			fmt.Println(r)
			result.Status = "Failed"
			result.NewEntries = 0
		}
	}()

	anime, err := api.GetAnimeByName(param.GetName())
	if err != nil {
		result.Status = "Failed"
		result.NewEntries = 0

		return result, fmt.Errorf("failed to fetch anime. %s", err.Error())
	}
	epInpChan := d.episodeStage(anime.GetID())
	go func() {
		loggingChan <- objects.Logging{
			Message: fmt.Sprintf("Found Anime %s", anime.GetTitle()),
			Level:   objects.L_INFO,
		}
	}()

	fmt.Printf("got anime %s\n", anime.GetTitle())

	dbInpChan := make(chan objects.DBRecords)

	go func() {
		dbInpChan <- anime
		close(dbInpChan)
	}()

	dbChan := insertObj(d.db, doneChan, dbInpChan, loggingChan)
	// close the channel returned from episodeStage
	epOutChan := insertObj(d.db, doneChan, epInpChan, loggingChan)

	for v := range utils.Multiplexer(dbChan, epOutChan) {
		fmt.Println("added record")
		result.NewEntries += int32(v)
	}

	// seed download list after anime and episode table has been populated
	result.NewEntries += insert(d.db, loggingChan, queryMap[DOWNLOAD_SEED], anime.GetID())

	return result, err
}

func NewDBServer(dbLogger *slog.Logger) *dbRPC {
	dbRpcSvc := new(dbRPC)
	dbRpcSvc.db = InitDB(dbLogger)
	dbRpcSvc.logger = dbLogger
	dbRpcSvc.httpClient = &http.Client{}
	dbRpcSvc.maxWorkers = 10

	client := os.Getenv("ANI_CLIENT")
	clientver := os.Getenv("ANI_CLIENTVER")
	protocol := os.Getenv("ANI_PROTOVER")
	request := os.Getenv("ANI_REQUEST")
	dbRpcSvc.dbLookupQuery = utils.Must(utils.AddParams(aniLookup, "client", client, "clientver", clientver, "protover", protocol, "request", request))

	return dbRpcSvc
}
