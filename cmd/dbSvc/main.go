package main

import (
	"log/slog"
	"net"
	"os"

	dbhandler "github.com/ayushsherpa111/anirss/cmd/dbSvc/dbHandler"
	"github.com/ayushsherpa111/anirss/pkg/rpc/dbservice"
	"github.com/ayushsherpa111/anirss/pkg/utils"
	"google.golang.org/grpc"
)

const (
	aniLog = "lookup.log"
	dbPort = 8282
)

var dbServerAddr = net.IP([]byte{0, 0, 0, 0})

func main() {
	dbLogger := utils.NewLogger(aniLog)
	defer func(logger *slog.Logger) {
		if r := recover(); r != nil {
			dbLogger.Debug(r.(string), "status", "recovered crash.")
		}
	}(dbLogger)
	utils.LoadEnv(dbLogger)
	dbSvcServer := dbhandler.NewDBServer(dbLogger)

	// setup TCP listener on port {dbPort}
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   dbServerAddr,
		Port: dbPort,
	})
	if err != nil {
		dbLogger.Error(err.Error())
		os.Exit(1)
	}

	// instantiate gRPC server
	grpcServer := grpc.NewServer()
	dbservice.RegisterAniDbSvcServer(grpcServer, dbSvcServer)
	dbLogger.Info("db grpc Server listening", "port", dbPort)
	if err := grpcServer.Serve(listener); err != nil {
		dbLogger.Error(err.Error())
		os.Exit(1)
	}
}
