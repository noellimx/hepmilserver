package statistics

import (
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
	postCh := reddit_miner.SubRedditPosts(subRedditName, postsCreatedWithinPast, reddit_miner.OrderByAlgoTop, true)

	var postForms []statisticsRepo.PostForm
	for p := range postCh {
		srName := strings.Replace(p.SubredditPrefixedName, "r/", "", -1)
		postForms = append(postForms, statisticsRepo.PostForm{
			Title:                         p.Title,
			PermaLinkPath:                 p.PermaLinkPath,
			DataKsId:                      p.DataKsId,
			Score:                         p.Score,
			SubredditId:                   p.SubredditId,
			CommentCount:                  p.CommentCount,
			SubredditName:                 srName,
			PolledTime:                    now,
			AuthorId:                      p.AuthorId,
			AuthorName:                    p.AuthorName,
			PolledTimeRoundedMinute:       roundDownTo5Mins,
			Rank:                          p.Rank,
			RankOrderType:                 statisticsRepo.OrderByAlgo(p.RankOrderType),
			RankOrderForCreatedWithinPast: statisticsRepo.CreatedWithinPast(p.RankOrderForCreatedWithinPast),
		})
	}
	//
	//log.Printf("PostForms: %#v\n", len(postForms))
	//log.Printf("Posts: %#v\n", len(posts))

	s.repo.InsertMany(postForms)
}
