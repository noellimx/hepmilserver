package statistics

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	conn *pgxpool.Pool
}

func NewAAA(conn *pgxpool.Pool) *Repo {
	return &Repo{
		conn: conn,
	}
}

type OrderByAlgo string

const (
	OrderByAlgoTop  OrderByAlgo = "top"
	OrderByAlgoBest OrderByAlgo = "best"
	OrderByAlgoHot  OrderByAlgo = "hot"
	OrderByAlgoNew  OrderByAlgo = "new"
)

type CreatedWithinPast string

const (
	CreatedWithinPastHour  CreatedWithinPast = "hour"
	CreatedWithinPastDay   CreatedWithinPast = "day"
	CreatedWithinPastMonth CreatedWithinPast = "month"
	CreatedWithinPastYear  CreatedWithinPast = "year"
)

type PostForm struct {
	Title                   string
	PermaLinkPath           string
	DataKsId                string
	Score                   *int32
	SubredditId             string
	CommentCount            *int32
	SubredditName           string
	PolledTime              time.Time
	AuthorId                string
	AuthorName              string
	PolledTimeRoundedMinute time.Time

	Rank                          int32
	RankOrderType                 OrderByAlgo
	RankOrderForCreatedWithinPast CreatedWithinPast
}

func (r *Repo) insert(post PostForm) error {
	row := r.conn.QueryRow(context.Background(), "insert into posts(title, perma_link_path, data_ks_id, score, subreddit_id, comment_count, subreddit_name, polled_time, author_id, author_name, polled_time_rounded_min, rank, rank_order_type, rank_order_created_within_past) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id",
		post.Title, post.PermaLinkPath, post.DataKsId, post.Score, post.SubredditId,
		post.CommentCount, post.SubredditName, post.PolledTime, post.AuthorId,
		post.AuthorName, post.PolledTimeRoundedMinute,
		post.Rank, post.RankOrderType, post.RankOrderForCreatedWithinPast,
	)
	var id int64
	return row.Scan(&id)
}

func (r *Repo) InsertMany(posts []PostForm) {
	log.Printf("InsertMany Posts: %#v\n", posts)
	for _, post := range posts {
		go func() {
			err := r.insert(post)
			if err != nil {
				log.Printf("InsertMany error %#v post %v\n", err, post)
			}
		}()
	}
}
