package main

import (
	"github.com/noellimx/hepmilserver/src/repository/task"
	"log"

	"github.com/noellimx/hepmilserver/src/infrastructure/stats_scraper"
)

func main() {
	posts, err := stats_scraper.SubRedditStatistics("memes", task.CreatedWithinPastDay, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("len(posts): %d %#v\n", len(posts), posts)
}
