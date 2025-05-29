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

type Interval = string

const (
	IntervalHour Interval = "hour"
)

type CreatedWithinPast = string

const (
	CreatedWithinPastHour  CreatedWithinPast = "hour"
	CreatedWithinPastDay   CreatedWithinPast = "day"
	CreatedWithinPastMonth CreatedWithinPast = "month"
	CreatedWithinPastYear  CreatedWithinPast = "year"
)

type OrderByColumn string

const (
	OrderByColumnTop  OrderByColumn = "top"
	OrderByColumnBest OrderByColumn = "best"
	OrderByColumnHot  OrderByColumn = "hot"
	OrderByColumnNew  OrderByColumn = "new"
)

func (r *Repo) Create(subRedditName string, itemCount int64, interval Interval, by OrderByColumn, itemsCreatedWithin CreatedWithinPast) error {
	row := r.conn.QueryRow(context.Background(), "insert into tasks(subreddit_name, min_item_count, interval, order_by, items_created_within_past) VALUES ($1,$2,$3,$4) RETURNING id", subRedditName, itemCount, interval, by, itemsCreatedWithin)
	var id int64
	return row.Scan(&id)
}

func (r *Repo) Delete(subRedditName string) error {
	_, err := r.conn.Exec(context.Background(), "DELETE FROM tasks where subreddit_name=$1", subRedditName)
	return err
}

func (r *Repo) GetByTaskInterval(every Interval) ([]string, error) {
	rows, err := r.conn.Query(context.Background(), "select subreddit_name, min_item_count, interval, order_by, items_created_within_past from tasks where interval = $1", every)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var subreddit_name string
		rows.Scan(&subreddit_name)
		if err := rows.Err(); err != nil {
			return []string{}, err
		}
		items = append(items, subreddit_name)
	}
	return items, nil
}
