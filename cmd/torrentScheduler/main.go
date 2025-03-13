package main

import (
	"fmt"
	"net"
	"os"

	"github.com/ayushsherpa111/anirss/pkg/utils"

	"google.golang.org/grpc"
)

/*
Triggered via Cron Job.
Get Episodes to be downloaded from the db service.
Fetch magnet URL from torrentURI/RSS feed
*/

const (
	torrentURI   = "https://nyaa.si/?page=rss&c=1_2&f=0" // add q=<anime_name>
	schedulerLog = "scheduler.log"
	svPORT       = 8383
)

var serverAddr = net.IP([]byte{0, 0, 0, 0})

func main() {
	schLogger := utils.NewLogger(schedulerLog)
	utils.LoadEnv(schLogger)

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   serverAddr,
		Port: svPORT,
	})
	if err != nil {
		schLogger.Error(fmt.Sprintf("failed to start TCP server. %s", err.Error()), "port", svPORT)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	schLogger.Info("gRPC server listening.")
	if err := grpcServer.Serve(listener); err != nil {
		schLogger.Error(err.Error())
		os.Exit(1)
	}
}
