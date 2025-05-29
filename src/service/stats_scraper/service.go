package stats_scraper

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

type Post struct {
	PermaLinkPath string `json:"perma_link_path"`
	DataKsId      string `json:"data_ks_id"`
	Score         string `json:"score"`
	SubId         string `json:"sub_id"`
	Title         string `json:"post_title"`
	CommentCount  string `json:"comment_count"`
}

type TimeFrame string

const (
	TimeFrame_Day TimeFrame = "day"
)

func SubRedditStatistics(subReddit string, timeFrame TimeFrame, debugEnabled bool) ([]Post, error) {
	if timeFrame != TimeFrame_Day {
		return []Post{}, fmt.Errorf("time frame not supported %s", timeFrame)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)

	// 2. Create an ExecAllocator with custom options
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	var ctxOpts []chromedp.ContextOption
	if debugEnabled {
		ctxOpts = append(ctxOpts, chromedp.WithDebugf(log.Printf))
	}
	ctx, cancel := chromedp.NewContext(
		allocCtx, ctxOpts...,
	)
	defer cancel()

	var articles []Post

	url := "https://www.reddit.com/r/" + subReddit
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, exp, err := runtime.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`).Do(ctx)
			time.Sleep(1 * time.Second)
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
		    const post_title = el.getAttribute('post-title');
		    const comment_count = el.getAttribute('comment-count');

		   return { perma_link_path,post_title,comment_count, data_ks_id, score, sub_id: el.getAttribute('subreddit-id')}
		})`, &articles),
	)

	if err != nil {
		return nil, fmt.Errorf("could not get articles: %w, url: %s", err, url)
	}

	return articles, nil
}
