package statistics

import (
	"fmt"
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

func (s Service) Scrape(subRedditName string, postsCreatedWithinPast reddit_miner.CreatedWithinPast, algo reddit_miner.OrderByAlgo) {
	now := time.Now().UTC()
	roundDownTo5Mins := now.Truncate(1 * time.Minute)
	postCh := reddit_miner.SubRedditPosts(subRedditName, postsCreatedWithinPast, algo, true)

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

func (s Service) Stats(name string, orderType statisticsRepo.OrderByAlgo, past statisticsRepo.CreatedWithinPast, granularity statisticsRepo.Granularity, fromTime *time.Time, toTime *time.Time) ([]statisticsRepo.Post, error) {
	if orderType != statisticsRepo.OrderByAlgoTop {
		return []statisticsRepo.Post{}, fmt.Errorf("order algo type %s not supported", orderType)
	}

	if past != statisticsRepo.CreatedWithinPastDay {
		return []statisticsRepo.Post{}, fmt.Errorf("past day %s not supported", past)
	}

	if name == "" {
		return []statisticsRepo.Post{}, fmt.Errorf("empty subreddit name")
	}

	switch granularity {
	case statisticsRepo.GranularityDaily, statisticsRepo.GranularityHour, statisticsRepo.GranularityQuarterHour:
		break
	default:
		return nil, fmt.Errorf("granularity type not supported. =%d", granularity)
	}

	posts, err := s.repo.Stats(name, orderType, fromTime, toTime, past, granularity)
	if err != nil {
		return []statisticsRepo.Post{}, err
	}

	return posts, nil

}
