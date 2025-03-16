set -x

echo "building database service"
go build -o dbService "./cmd/dbSvc/main.go"

echo "building torrent scheduler"
go build -o scheduler "./cmd/torrentScheduler/main.go"
