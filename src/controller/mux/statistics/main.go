package statistics

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/noellimx/hepmilserver/src/controller/response_types"
	"github.com/noellimx/hepmilserver/src/httplog"

	statisticsrepo "github.com/noellimx/hepmilserver/src/infrastructure/repositories/statistics"
	statisticsservice "github.com/noellimx/hepmilserver/src/service/statistics"
)

type Handlers struct {
	service *statisticsservice.Service
}

func NewHandlers(service *statisticsservice.Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

// Get godoc
// @Summary      Retrieve time series data in denormalized form.
// @Description  Retrieve time series data in denormalized form.
// @Tags         subreddit
// @Param        subreddit_name   					query      string  true  "name"
// @Param        rank_order_type   					query      string  true  "[top,best,hot,new]"
// @Param        rank_order_created_within_past   	query      string  true  "[hour,day,month,year]"
// @Param        granularity   						query      string  true  "1=Minute,2=QuarterHour,3=Hour,4=Daily,5=Monthly"
// @Param        backfill   						query      string  true  "true=backfill incomplete data"
// @Accept       json, text/csv
// @Produce      json, text/csv
// @Success      200  {object}  GetStatisticsResponseBody
// @Failure      500  {object}  ErrorResponse
// @Router       /statistics [get]
func (h Handlers) Get(w http.ResponseWriter, r *http.Request) {
	prefix := httplog.SPrintHttpRequestPrefix(r)

	_subRedditName := r.URL.Query().Get("subreddit_name")
	_rankOrderType := r.URL.Query().Get("rank_order_type")
	_rankOrderCreatedWithinPast := r.URL.Query().Get("rank_order_created_within_past")
	_granularity := r.URL.Query().Get("granularity")
	_fromTime := r.URL.Query().Get("from_time")
	_toTime := r.URL.Query().Get("to_time")
	_shouldBackfill := r.URL.Query().Get("backfill")

	contentType := r.Header.Get("Accept")

	if contentType != "application/json" && contentType != "text/csv" {
		response_types.ErrorNoBody(w, http.StatusUnsupportedMediaType, fmt.Errorf("content type %s not supported", contentType))
		return
	}

	fromTime, err := time.Parse("2006-01-02T15:04:05.000Z", _fromTime)
	if err != nil {
		response_types.ErrorNoBody(w, http.StatusBadRequest, err)
		return
	}

	toTime, err := time.Parse("2006-01-02T15:04:05.000Z", _toTime)
	if err != nil {
		response_types.ErrorNoBody(w, http.StatusBadRequest, err)
		return
	}
	fmt.Printf("TIME from %s to %s \n", fromTime, toTime)

	backfill := _shouldBackfill == "true"
	switch "" {
	case _subRedditName, _rankOrderType, _rankOrderCreatedWithinPast, _granularity:
		err := fmt.Errorf("some field is empty. subreddit_name %v, rank_order_type %v, rank_order_created_within_past %v, granularity %v", _subRedditName, _rankOrderType, _rankOrderCreatedWithinPast, _granularity)
		log.Printf("%s error=%v\n", prefix, err)
		response_types.ErrorNoBody(w, http.StatusBadRequest, err)
		return
	}

	granularity, _ := strconv.Atoi(_granularity)

	posts, err := h.service.Stats(_subRedditName, statisticsrepo.OrderByAlgo(_rankOrderType), statisticsrepo.CreatedWithinPast(_rankOrderCreatedWithinPast), statisticsrepo.Granularity(granularity), &fromTime, &toTime, backfill)
	if err != nil {
		log.Printf("%s error=%v\n", prefix, err)
		response_types.ErrorNoBody(w, http.StatusBadRequest, err)
		return
	}

	switch contentType {
	case "application/json":
		response_types.OkJsonBody(w, GetStatisticsResponseBodyData{
			Posts: toJSON(posts),
		})
	case "text/csv":
		layout := "2006-01-02_15-04-05"
		ftString := fromTime.Format(layout)
		ttString := toTime.Format(layout)

		csvName := fmt.Sprintf(`%s_%s_%s_FROM_%s_TO_%s`, _subRedditName, _rankOrderType, _rankOrderCreatedWithinPast, ftString, ttString)
		response_types.Csv(w, csvName, toCSV(posts))
	}
}

func toCSV(posts []statisticsservice.Post) [][]string {
	header := []string{
		"row_position",

		"polled_time",
		"polled_time_rounded_min",

		"rank",
		"rank_order_type",
		"rank_order_created_within_past",

		"data_ks_id",
		"title",
		"perma_link_path",

		"comment_count",
		"score",
		"subreddit_id",
		"subreddit_name",
		"author_id",
		"author_name",
	}

	rows := [][]string{header}

	for i, p := range posts {
		var score string
		if p.Score != nil {
			score = strconv.FormatInt(int64(*p.Score), 10)
		}

		var commentCount string
		if p.CommentCount != nil {
			commentCount = strconv.FormatInt(int64(*p.CommentCount), 10)
		}

		var rank string
		if p.Rank != nil {
			rank = strconv.FormatInt(int64(*p.Rank), 10)
		}

		rows = append(rows, []string{
			strconv.Itoa(i),

			p.PolledTime.UTC().String(),
			p.PolledTimeRoundedMinute.UTC().String(),

			rank,
			string(p.RankOrderType),
			string(p.RankOrderForCreatedWithinPast),

			p.DataKsId,
			p.Title,
			p.PermaLinkPath,

			commentCount,
			score,

			p.SubredditId,
			p.SubredditName,

			p.AuthorId,
			p.AuthorName,
		})
	}
	return rows
}

func toJSON(posts []statisticsservice.Post) (ps []Post) {
	for _, post := range posts {
		ps = append(ps, Post{
			Title:                         post.Title,
			PermaLinkPath:                 post.PermaLinkPath,
			DataKsId:                      post.DataKsId,
			Score:                         post.Score,
			SubredditId:                   post.SubredditId,
			CommentCount:                  post.CommentCount,
			SubredditName:                 post.SubredditName,
			PolledTime:                    post.PolledTime,
			AuthorId:                      post.AuthorId,
			AuthorName:                    post.AuthorName,
			PolledTimeRoundedMinute:       post.PolledTimeRoundedMinute,
			Rank:                          post.Rank,
			RankOrderType:                 post.RankOrderType,
			RankOrderForCreatedWithinPast: post.RankOrderForCreatedWithinPast,
			IsSynthetic:                   post.IsSynthetic,
		})
	}
	return
}

type Post struct {
	Title         string `json:"title"`
	PermaLinkPath string `json:"perma_link_path"`
	DataKsId      string `json:"data_ks_id"`
	Score         *int32 `json:"score"`
	SubredditId   string `json:"subreddit_id"`

	CommentCount  *int32    `json:"comment_count"`
	SubredditName string    `json:"subreddit_name"`
	PolledTime    time.Time `json:"polled_time"`
	AuthorId      string    `json:"author_id"`
	AuthorName    string    `json:"author_name"`

	PolledTimeRoundedMinute       time.Time                        `json:"polled_time_rounded_min"`
	Rank                          *int32                           `json:"rank"`
	RankOrderType                 statisticsrepo.OrderByAlgo       `json:"rank_order_type"`
	RankOrderForCreatedWithinPast statisticsrepo.CreatedWithinPast `json:"rank_order_created_within_past"`

	IsSynthetic bool `json:"is_synthetic"`
}

type GetStatisticsResponseBodyData struct {
	Posts []Post `json:"posts"`
}
type GetStatisticsResponseBody = response_types.Response[GetStatisticsResponseBodyData]
type ErrorResponse = response_types.Response[struct{}]
