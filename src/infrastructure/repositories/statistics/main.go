package statistics

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

type Repo struct {
	conn *pgxpool.Pool
}

func NewAAA(conn *pgxpool.Pool) *Repo {
	return &Repo{
		conn: conn,
	}
}

type PostForm struct {
	Id                        int64
	Title                     string
	PermaLinkPath             string
	DataKsId                  string
	Score                     *int32
	SubredditId               string
	CommentCount              *int32
	SubredditName             string
	PolledTime                time.Time
	AuthorId                  string
	AuthorName                string
	PolledTimeRounded5Minutes time.Time
}

func (r *Repo) insert(post PostForm) error {
	row := r.conn.QueryRow(context.Background(), "insert into posts(title, perma_link_path, data_ks_id, score, subreddit_id, comment_count, subreddit_name, polled_time, author_id, author_name, polled_time_rounded_5min) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id", post.Title, post.PermaLinkPath, post.DataKsId, post.Score, post.SubredditId, post.CommentCount, post.SubredditName, post.PolledTime, post.AuthorId, post.AuthorName, post.PolledTimeRounded5Minutes)
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
