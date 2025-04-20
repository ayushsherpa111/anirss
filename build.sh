set -x

echo "building database service"
go build -o dbService "./cmd/dbSvc/main.go"

echo "building torrent service"
go build -o scheduler "./cmd/torrentScheduler/main.go"

echo "building client service"
go build -o client "./cmd/clientSvc/main.go"
