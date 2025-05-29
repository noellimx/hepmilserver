package main

import (
	"log"

	"github.com/noellimx/hepmilserver/src/service/stats_scraper"
)

func main() {
	posts, err := stats_scraper.SubRedditStatistics("memes", stats_scraper.TimeFrame_Day, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("len(posts): %d %#v\n", len(posts), posts)
}
