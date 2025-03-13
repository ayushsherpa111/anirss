package main

// "github.com/mmcdole/gofeed"
import (
	"fmt"
)

const (
	RSS = "https://nyaa.si/?page=rss" // RSS feed for new episodes
// http://api.anidb.net:9001/httpapi\?client\=anirsslookup\&clientver\=1\&protover\=1\&request\=anime\&aid\=18238
)

func main() {
	fmt.Println("hello")
}

// logFile, err := os.OpenFile("anirss.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
// if err != nil {
//   log.Fatalf("Failed to open log file. ERR: %s", err.Error())
// }
// defer logFile.Close()
//
// logger := log.New(logFile, "[ANIRSS]", log.Lshortfile|log.LUTC|log.LstdFlags)
// logger.Println("Initialized logger.")
//
// rssURI, err := url.Parse(RSS)
// if err != nil {
//   logger.Fatalf("failed to parse rss URL. Recheck the RSS URI. ERR: %s\n", err.Error())
// }
//
// queryVals := rssURI.Query()
// queryVals.Set("q", "dragon ball daima")
// rssURI.RawQuery = queryVals.Encode()
// fmt.Println(rssURI.String())
//
