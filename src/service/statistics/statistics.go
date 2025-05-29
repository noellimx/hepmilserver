package statistics

import (
	"strconv"
	"strings"
	"time"

	"github.com/noellimx/hepmilserver/src/infrastructure/reddit_miner"
	statisticsRepo "github.com/noellimx/hepmilserver/src/infrastructure/repositories/statistics"
)

type Service struct {
	repo *statisticsRepo.Repo
}

func NewWWW(repo *statisticsRepo.Repo) *Service {
	return &Service{repo: repo}
}

func (s Service) Scrape(subRedditName string, postsCreatedWithinPast reddit_miner.CreatedWithinPast) {
	now := time.Now().UTC()
	roundDownTo5Mins := now.Truncate(1 * time.Minute)
	postCh := reddit_miner.SubRedditPosts(subRedditName, postsCreatedWithinPast, reddit_miner.OrderByColumnTop, true)

	var postForms []statisticsRepo.PostForm
	for p := range postCh {
		var commentCount *int32
		_commentCount, err := strconv.Atoi(p.CommentCount)
		if err == nil {
			c := int32(_commentCount)
			commentCount = &c
		}

		var score *int32
		_score, err := strconv.Atoi(p.Score)
		if err == nil {
			c := int32(_score)
			score = &c
		}
		srName := strings.Replace(p.SubredditPrefixedName, "r/", "", -1)
		postForms = append(postForms, statisticsRepo.PostForm{
			Title:                     p.Title,
			PermaLinkPath:             p.PermaLinkPath,
			DataKsId:                  p.DataKsId,
			Score:                     score,
			SubredditId:               p.SubredditId,
			CommentCount:              commentCount,
			SubredditName:             srName,
			PolledTime:                now,
			AuthorId:                  p.AuthorId,
			AuthorName:                p.AuthorName,
			PolledTimeRounded5Minutes: roundDownTo5Mins,
		})
	}
	//
	//log.Printf("PostForms: %#v\n", len(postForms))
	//log.Printf("Posts: %#v\n", len(posts))

	s.repo.InsertMany(postForms)
}
