package main

// "github.com/mmcdole/gofeed"
import (
	"context"
	"fmt"

	"github.com/ayushsherpa111/anirss/pkg/rpc/dbservice"
	"github.com/ayushsherpa111/anirss/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	dbService = "localhost:8282"
	clientLog = "client.log"
)

func main() {
	clientLogger := utils.NewLogger(clientLog)
	client, err := grpc.NewClient(dbService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		clientLogger.Error("Failed to create new grpc Client", "err", err.Error())
		panic(err)
	}
	dbSvc := dbservice.NewAniDbSvcClient(client)
	result, err := dbSvc.AddAnimeByName(context.Background(), &dbservice.AniParams{
		Name: "Attack on titan",
	})
	if err != nil {
		clientLogger.Error(err.Error())
	}
	clientLogger.Info("AddAnime result", "result", result)
	fmt.Println(result, err)
}
