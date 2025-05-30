package task

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	conn *pgxpool.Pool
}

func New(conn *pgxpool.Pool) *Repo {
	return &Repo{
		conn: conn,
	}
}

type Interval string

const (
	IntervalHour Interval = "hour"
)

type CreatedWithinPast string

const (
	CreatedWithinPastHour  CreatedWithinPast = "hour"
	CreatedWithinPastDay   CreatedWithinPast = "day"
	CreatedWithinPastMonth CreatedWithinPast = "month"
	CreatedWithinPastYear  CreatedWithinPast = "year"
)

type OrderByAlgo string

const (
	OrderByAlgoTop  OrderByAlgo = "top"
	OrderByAlgoBest OrderByAlgo = "best"
	OrderByAlgoHot  OrderByAlgo = "hot"
	OrderByAlgoNew  OrderByAlgo = "new"
)

func (r *Repo) Create(subRedditName string, itemCount int64, interval Interval, by OrderByAlgo, itemsCreatedWithin CreatedWithinPast) error {
	row := r.conn.QueryRow(context.Background(), "insert into tasks(subreddit_name, min_item_count, interval, order_by, items_created_within_past) VALUES ($1,$2,$3,$4, $5) RETURNING id", subRedditName, itemCount, interval, by, itemsCreatedWithin)
	var id int64
	return row.Scan(&id)
}

func (r *Repo) Delete(subRedditName string) error {
	_, err := r.conn.Exec(context.Background(), "DELETE FROM tasks where subreddit_name=$1", subRedditName)
	return err
}

type Task struct {
	Id                     int64
	SubRedditName          string
	MinItemCount           int64
	Interval               Interval
	OrderBy                OrderByAlgo
	PostsCreatedWithinPast CreatedWithinPast
}

func (r *Repo) GetByTaskInterval(every Interval) ([]Task, error) {
	rows, err := r.conn.Query(context.Background(), "select id, subreddit_name, min_item_count, interval, order_by, items_created_within_past from tasks where interval = $1", every)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var task []Task
	for rows.Next() {
		var t Task
		rows.Scan(&t.Id, &t.SubRedditName, &t.MinItemCount, &t.Interval, &t.OrderBy, &t.PostsCreatedWithinPast)
		if err := rows.Err(); err != nil {
			return []Task{}, err
		}
		task = append(task, t)
	}
	return task, nil
}
