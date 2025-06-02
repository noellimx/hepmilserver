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

type Granularity int64

const (
	_                      Granularity = 0
	GranularityMinute      Granularity = 1
	GranularityQuarterHour Granularity = 2
	GranularityHour        Granularity = 3
	GranularityDaily       Granularity = 4
)

type PostForm struct {
	Title                   string
	SubredditName           string
	PolledTime              time.Time
	PolledTimeRoundedMinute time.Time

	CommentCount *int32
	Score        *int32

	Rank                          int32
	RankOrderType                 OrderByAlgo
	RankOrderForCreatedWithinPast CreatedWithinPast

	SubredditId   string
	DataKsId      string
	PermaLinkPath string

	AuthorId   string
	AuthorName string
}

func (r *Repo) insert(post PostForm) error {
	row := r.conn.QueryRow(context.Background(), "insert into post_statistics(title, perma_link_path, data_ks_id, score, subreddit_id, comment_count, subreddit_name, polled_time, author_id, author_name, polled_time_rounded_min, rank, rank_order_type, rank_order_created_within_past) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id",
		post.Title, post.PermaLinkPath, post.DataKsId, post.Score, post.SubredditId,
		post.CommentCount, post.SubredditName, post.PolledTime, post.AuthorId,
		post.AuthorName, post.PolledTimeRoundedMinute,
		post.Rank, post.RankOrderType, post.RankOrderForCreatedWithinPast,
	)
	var id int64
	return row.Scan(&id)
}

func (r *Repo) InsertMany(posts []PostForm) {
	log.Printf("InsertMany Posts length: %d\n", len(posts))
	for _, post := range posts {
		go func() {
			err := r.insert(post)
			if err != nil {
				log.Printf("InsertMany error %#v post %v\n", err, post)
			}
		}()
	}
}

type Post struct {
	Title                         string
	PermaLinkPath                 string
	DataKsId                      string
	Score                         int32
	SubredditId                   string
	CommentCount                  int32
	SubredditName                 string
	PolledTime                    time.Time
	AuthorId                      string
	AuthorName                    string
	PolledTimeRoundedMinute       time.Time
	Rank                          int32
	RankOrderType                 OrderByAlgo
	RankOrderForCreatedWithinPast CreatedWithinPast
	Id                            int64
}

func (r *Repo) Stats(name string, orderType OrderByAlgo, fromTime *time.Time, toTime *time.Time, past CreatedWithinPast, granularity Granularity) ([]Post, error) {
	rows, err := r.conn.Query(context.Background(), `select id,
		title,
		perma_link_path,
		data_ks_id,
		score,
		subreddit_id,
		
		comment_count,
		subreddit_name,
		polled_time,
		author_id,
		author_name,
		
		polled_time_rounded_min,
		rank,
		rank_order_type,
		rank_order_created_within_past
		from post_statistics
		where true
		and rank <= 20
		and subreddit_name = $1
		and rank_order_type = $2
		and rank_order_created_within_past = $3
		and extract(minute from polled_time_rounded_min)::integer % 60 = 0
		and $4 < polled_time_rounded_min
		and polled_time_rounded_min < $5  
;`, name, orderType, past, fromTime, toTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var post []Post
	for rows.Next() {
		var t Post
		rows.Scan(&t.Id, &t.Title, &t.PermaLinkPath, &t.DataKsId, &t.Score, &t.SubredditId,
			&t.CommentCount, &t.SubredditName, &t.PolledTime,
			&t.AuthorId,
			&t.AuthorName,
			&t.PolledTimeRoundedMinute,
			&t.Rank,
			&t.RankOrderType,
			&t.RankOrderForCreatedWithinPast,
		)
		if err := rows.Err(); err != nil {
			return []Post{}, err
		}
		post = append(post, t)
	}
	return post, nil

}
