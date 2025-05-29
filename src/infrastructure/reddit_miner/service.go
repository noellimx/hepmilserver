package reddit_miner

import (
	"context"
	"fmt"
	"log"
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

type Post struct {
	Title         string `json:"title"`
	PermaLinkPath string `json:"perma_link_path"`
	DataKsId      string `json:"data_ks_id"` // the raw post id prepended with `t3_`, i.e t3_1ky2rld

	SubredditId           string `json:"subreddit_id"`
	SubredditPrefixedName string `json:"subreddit_prefix_name"`

	Score        string `json:"score"`
	CommentCount string `json:"comment_count"`

	AuthorId   string `json:"author_id"`
	AuthorName string `json:"author"`

	CreatedTimestamp string `json:"created_timestamp"`
}

type OrderByColumn string

const (
	OrderByColumnTop  OrderByColumn = "top"
	OrderByColumnBest OrderByColumn = "best"
	OrderByColumnHot  OrderByColumn = "hot"
	OrderByColumnNew  OrderByColumn = "new"
)

func SubRedditPosts(subReddit string, createdWithinPast CreatedWithinPast, orderBy OrderByColumn, debugLogEnabled bool) <-chan Post {
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

		var posts []Post
		chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.ActionFunc(func(ctx context.Context) error {
				log.Printf("action func")
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
			chromedp.Evaluate(`Array.from(document.querySelectorAll("[data-ks-item]")).map(el => {
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
		
		   return { subreddit_id, subreddit_prefix_name, perma_link_path,title,comment_count, data_ks_id, score, created_timestamp, author_id, author }
		})`, &posts),
		)
		for _, post := range posts {
			ch <- post
		}

		close(ch)
	}()
	return ch
}
