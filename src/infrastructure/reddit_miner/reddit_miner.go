package reddit_miner

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type CreatedWithinPast string

const (
	CreatedWithinPastHour  CreatedWithinPast = "hour"
	CreatedWithinPastDay   CreatedWithinPast = "day"
	CreatedWithinPastMonth CreatedWithinPast = "month"
	CreatedWithinPastYear  CreatedWithinPast = "year"
)

type PostDom struct {
	Title         string `json:"title"`
	DataKsId      string `json:"data_ks_id"` // the raw post id prepended with `t3_`, i.e t3_1ky2rld
	PermaLinkPath string `json:"perma_link_path"`

	SubredditId           string `json:"subreddit_id"`
	SubredditPrefixedName string `json:"subreddit_prefix_name"`

	AuthorId   string `json:"author_id"`
	AuthorName string `json:"author"`

	CreatedTimestamp string `json:"created_timestamp"`

	Score        string `json:"score"`
	CommentCount string `json:"comment_count"`

	Index int32 `json:"index"`
}

type Post struct {
	Title         string
	DataKsId      string
	PermaLinkPath string

	SubredditId           string
	SubredditPrefixedName string

	AuthorId   string
	AuthorName string

	CreatedTimestamp string

	Score        *int32
	CommentCount *int32

	Rank                          int32
	RankOrderType                 OrderByAlgo
	RankOrderForCreatedWithinPast CreatedWithinPast
}

type OrderByAlgo string

const (
	OrderByAlgoTop  OrderByAlgo = "top"
	OrderByAlgoBest OrderByAlgo = "best"
	OrderByAlgoHot  OrderByAlgo = "hot"
	OrderByAlgoNew  OrderByAlgo = "new"
)

func SubRedditPosts(subReddit string, createdWithinPast CreatedWithinPast, orderBy OrderByAlgo, debugLogEnabled bool) <-chan Post {
	ch := make(chan Post)
	go func() {
		if createdWithinPast != CreatedWithinPastDay {
			log.Printf("time frame not supported %s", createdWithinPast)
			return
		}

		if subReddit == "" {
			log.Printf("SubRedditPosts() subReddit Name is empty")
			return
		}
		if orderBy == "" {
			log.Printf("SubRedditPosts() subReddit Name is empty")
			return
		}

		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
		)

		// 2. Create an ExecAllocator with custom options
		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancel()

		var ctxOpts []chromedp.ContextOption
		if debugLogEnabled {
			ctxOpts = append(ctxOpts, chromedp.WithDebugf(log.Printf))
		}
		ctx, cancel := chromedp.NewContext(
			allocCtx, ctxOpts...,
		)
		defer cancel()

		url := fmt.Sprintf("https://www.reddit.com/r/%s/%s?t=%s", subReddit, orderBy, createdWithinPast)
		log.Printf("SubRedditPosts() URL: %s", url)

		var posts []PostDom
		chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.ActionFunc(func(ctx context.Context) error {
				_, exp, err := runtime.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`).Do(ctx)
				time.Sleep(10 * time.Second)
				if err != nil {
					return err
				}
				if exp != nil {
					return exp
				}
				return nil
			}),
			chromedp.Evaluate(`Array.from(document.querySelectorAll("[data-ks-item]")).map((el, index) => {
		    const data_ks_id = el.querySelector("a").getAttribute('data-ks-id');
		    const perma_link_path = el.getAttribute('permalink');
		    const score = el.getAttribute('score');
		    const title = el.getAttribute('post-title');
		    const comment_count = el.getAttribute('comment-count');
		    const subreddit_id = el.getAttribute('subreddit-id');
		    const subreddit_prefix_name = el.getAttribute('subreddit-prefixed-name');
		    const created_timestamp = el.getAttribute('created-timestamp');
		    const author_id = el.getAttribute('author-id');
		    const author = el.getAttribute('author');
		
		   return { index, subreddit_id, subreddit_prefix_name, perma_link_path,title,comment_count, data_ks_id, score, created_timestamp, author_id, author }
		})`, &posts),
		)
		for _, p := range posts {
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
			log.Printf("p.Rank %v", p.Index)

			ch <- Post{
				Title:                         p.Title,
				DataKsId:                      p.DataKsId,
				PermaLinkPath:                 p.PermaLinkPath,
				SubredditId:                   p.SubredditId,
				SubredditPrefixedName:         p.SubredditPrefixedName,
				AuthorId:                      p.AuthorId,
				AuthorName:                    p.AuthorName,
				CreatedTimestamp:              p.CreatedTimestamp,
				Score:                         score,
				CommentCount:                  commentCount,
				Rank:                          p.Index + 1,
				RankOrderType:                 orderBy,
				RankOrderForCreatedWithinPast: createdWithinPast,
			}
		}
		close(ch)
	}()
	return ch
}
