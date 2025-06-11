package main

import (
	"log"
	"time"

	"github.com/noellimx/redditminer/src/infrastructure/reddit_miner"
)

func main() {
	postsC := reddit_miner.SubRedditPosts("memes", reddit_miner.CreatedWithinPastDay, reddit_miner.OrderByAlgoTop, false)
	var posts []reddit_miner.Post
	for p := range postsC {
		posts = append(posts, p)
	}
	time.Sleep(10 * time.Second)

	log.Printf("len(posts): %d %#v\n", len(posts), posts)

	time.Sleep(10 * time.Second)
}
