package statistics

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/noellimx/hepmilserver/src/infrastructure/reddit_miner"
	statisticsrepo "github.com/noellimx/hepmilserver/src/infrastructure/repositories/statistics"
)

type Service struct {
	repo *statisticsrepo.Repo
}

func NewWWW(repo *statisticsrepo.Repo) *Service {
	return &Service{repo: repo}
}

func (s Service) Scrape(subRedditName string, postsCreatedWithinPast reddit_miner.CreatedWithinPast, algo reddit_miner.OrderByAlgo) {
	now := time.Now().UTC()
	roundDownTo5Mins := now.Truncate(1 * time.Minute)
	postCh := reddit_miner.SubRedditPosts(subRedditName, postsCreatedWithinPast, algo, false)

	var postForms []statisticsrepo.PostForm
	for p := range postCh {
		srName := strings.Replace(p.SubredditPrefixedName, "r/", "", -1)
		postForms = append(postForms, statisticsrepo.PostForm{
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
			RankOrderType:                 statisticsrepo.OrderByAlgo(p.RankOrderType),
			RankOrderForCreatedWithinPast: statisticsrepo.CreatedWithinPast(p.RankOrderForCreatedWithinPast),
		})
	}
	//
	//log.Printf("PostForms: %#v\n", len(postForms))
	//log.Printf("Posts: %#v\n", len(posts))

	s.repo.InsertMany(postForms)
}

type Post struct {
	Title         string
	PermaLinkPath string

	DataKsId      string
	SubredditId   string
	SubredditName string
	AuthorId      string
	AuthorName    string

	PolledTime              time.Time
	PolledTimeRoundedMinute time.Time

	Score        *int32
	CommentCount *int32

	RankOrderType                 statisticsrepo.OrderByAlgo
	RankOrderForCreatedWithinPast statisticsrepo.CreatedWithinPast
	Rank                          *int32
	IsSynthetic                   bool
}

func minTimeF(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTimeF(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

var GranularityToDuration = map[statisticsrepo.Granularity]time.Duration{
	statisticsrepo.GranularityHour: time.Hour,
}

// Stats
// We will cleanse the data here
// - Dedupe
// - Backfill (#data points = #post x #time)
//   - incomplete series
//   - missing series
func (s Service) Stats(name string, orderType statisticsrepo.OrderByAlgo, past statisticsrepo.CreatedWithinPast, granularity statisticsrepo.Granularity, fromTime *time.Time, toTime *time.Time, shouldBackFill bool) ([]Post, error) {
	if orderType != statisticsrepo.OrderByAlgoTop {
		return []Post{}, fmt.Errorf("order algo type %s not supported", orderType)
	}

	if past != statisticsrepo.CreatedWithinPastDay {
		return []Post{}, fmt.Errorf("past day %s not supported", past)
	}

	if name == "" {
		return []Post{}, fmt.Errorf("empty subreddit name")
	}

	switch granularity {
	case statisticsrepo.GranularityHour:
		break
	default:
		return nil, fmt.Errorf("granularity type not supported. =%d", granularity)
	}

	postsDb, err := s.repo.Stats(name, orderType, fromTime, toTime, past, granularity)
	if err != nil {
		return []Post{}, err
	}

	if len(postsDb) == 0 {
		return []Post{}, nil
	}

	if !shouldBackFill {
		var posts []Post
		for _, post := range postsDb {
			posts = append(posts, cloneType(post))
		}

		slices.SortFunc(posts, func(a, b Post) int {
			if a.PolledTimeRoundedMinute.Before(b.PolledTimeRoundedMinute) {
				return -1
			}

			if b.PolledTimeRoundedMinute.Before(a.PolledTimeRoundedMinute) {
				return 1
			}
			if a.Rank != nil && b.Rank != nil {
				if *a.Rank < *b.Rank {
					return -1
				}
				if *b.Rank < *a.Rank {
					return 1
				}
			} else {
				log.Printf("?? why some nil")
			}
			return 0
		})
		return posts, nil
	}

	// cleansing..
	tick := GranularityToDuration[granularity]
	if tick == 0 {
		return []Post{}, fmt.Errorf("unknown converstion from granularity type to tick. =%d", granularity)
	}

	// build series by ks_id
	twoD := make(map[ /*ks_id*/ string]map[ /*polled time rounded */ time.Time]statisticsrepo.Post)

	allTimes := make(map[time.Time]struct{})
	minTime, maxTime := postsDb[0].PolledTimeRoundedMinute, postsDb[0].PolledTimeRoundedMinute

	// dedup
	for _, post := range postsDb {
		// build 2D
		m, ok := twoD[post.DataKsId]
		if !ok {
			m = make(map[time.Time]statisticsrepo.Post)
		}
		m[post.PolledTimeRoundedMinute] = post
		twoD[post.DataKsId] = m

		allTimes[post.PolledTimeRoundedMinute] = struct{}{}
		minTime = minTimeF(post.PolledTimeRoundedMinute, minTime)
		maxTime = maxTimeF(post.PolledTimeRoundedMinute, maxTime)
	}

	// backfill incomplete series
	var posts []Post
	for _, _posts := range twoD {
		firstTime := time.Time{}
		for k, _ := range _posts {
			firstTime = k
			break
		}
		firstPost := _posts[firstTime]
		for needTime, _ := range allTimes {
			p, ok := _posts[needTime]
			var _p Post
			if ok {
				_p = cloneType(p)
			} else {
				_p = Post{
					Title:         firstPost.Title,
					PermaLinkPath: firstPost.PermaLinkPath,
					DataKsId:      firstPost.DataKsId,
					SubredditId:   firstPost.SubredditId,
					Score:         nil,
					CommentCount:  nil,
					SubredditName: firstPost.SubredditName,
					AuthorId:      firstPost.AuthorId,
					AuthorName:    firstPost.AuthorName,

					PolledTime:              needTime,
					PolledTimeRoundedMinute: needTime,

					Rank:                          nil,
					RankOrderType:                 firstPost.RankOrderType,
					RankOrderForCreatedWithinPast: firstPost.RankOrderForCreatedWithinPast,
					IsSynthetic:                   true,
				}
			}
			posts = append(posts, _p)
		}
	}

	// backfill missing series
	for t := minTime; t.Before(maxTime); t = t.Add(tick) {
		_, inAllTimes := allTimes[t]
		if inAllTimes {
			continue
		}

		for ksId, _posts := range twoD {
			firstTime := time.Time{}
			for k, _ := range _posts {
				firstTime = k
				break
			}
			firstPost := _posts[firstTime]

			posts = append(posts, Post{
				Title:       firstPost.Title,
				DataKsId:    ksId,
				IsSynthetic: true,

				PermaLinkPath: firstPost.PermaLinkPath,
				SubredditId:   "",
				SubredditName: "",
				AuthorId:      "",
				AuthorName:    "",

				PolledTimeRoundedMinute: t,

				Score:                         nil,
				CommentCount:                  nil,
				RankOrderType:                 "",
				RankOrderForCreatedWithinPast: "",
				Rank:                          nil,
			})
		}
	}

	slices.SortFunc(posts, func(a, b Post) int {
		if a.PolledTimeRoundedMinute.Before(b.PolledTimeRoundedMinute) {
			return -1
		}

		if b.PolledTimeRoundedMinute.Before(a.PolledTimeRoundedMinute) {
			return 1
		}
		if a.Rank != nil && b.Rank != nil {
			if *a.Rank < *b.Rank {
				return -1
			}
			if *b.Rank < *a.Rank {
				return 1
			}
		} else {
			log.Printf("?? why some nil")
		}
		return 0
	})
	return posts, nil
}

func cloneType(p statisticsrepo.Post) Post {
	return Post{
		Title:                         p.Title,
		PermaLinkPath:                 p.PermaLinkPath,
		DataKsId:                      p.DataKsId,
		SubredditId:                   p.SubredditId,
		SubredditName:                 p.SubredditName,
		AuthorId:                      p.AuthorId,
		AuthorName:                    p.AuthorName,
		PolledTime:                    p.PolledTime,
		PolledTimeRoundedMinute:       p.PolledTimeRoundedMinute,
		Score:                         &p.Score,
		CommentCount:                  &p.CommentCount,
		RankOrderType:                 p.RankOrderType,
		RankOrderForCreatedWithinPast: p.RankOrderForCreatedWithinPast,
		Rank:                          &p.Rank,
		IsSynthetic:                   false,
	}
}
