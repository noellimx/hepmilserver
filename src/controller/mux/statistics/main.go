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
// @Param        rank_order_type   					query      string  true  "["top","best","hot","new"]"
// @Param        rank_order_created_within_past   	query      string  true  "["hour","day","month","year"]"
// @Param        granularity   						query      string  true  "1=Minute,2=QuarterHour,3=Hour,4=Daily,5=Mins"
// @Accept       json
// @Produce      json
// @Success      200  {object}  GetStatisticsResponseBody
// @Failure      500  {object}  ErrorResponse
// @Router       /statistics [get]
func (h Handlers) Get(w http.ResponseWriter, r *http.Request) {
	prefix := httplog.SPrintHttpRequestPrefix(r)
	_subRedditName := r.URL.Query().Get("subreddit_name")
	_rankOrderType := r.URL.Query().Get("rank_order_type")
	_rankOrderCreatedWithinPast := r.URL.Query().Get("rank_order_created_within_past")
	_granularity := r.URL.Query().Get("granularity")

	switch "" {
	case _subRedditName, _rankOrderType, _rankOrderCreatedWithinPast, _granularity:
		err := fmt.Errorf("some field is empty. subreddit_name %v, rank_order_type %v, rank_order_created_within_past %v, granularity %v", _subRedditName, _rankOrderType, _rankOrderCreatedWithinPast, _granularity)
		log.Printf("%s error=%v\n", prefix, err)
		response_types.ErrorNoBody(w, http.StatusBadRequest, err)
		return
	}

	granularity, _ := strconv.Atoi(_granularity)
	to := time.Now()
	from := to.Add(-24 * time.Hour)

	posts, err := h.service.Stats(_subRedditName, statisticsrepo.OrderByAlgo(_rankOrderType), statisticsrepo.CreatedWithinPast(_rankOrderCreatedWithinPast), statisticsrepo.Granularity(granularity), &to, &from)
	if err != nil {
		log.Printf("%s error=%v\n", prefix, err)
		response_types.ErrorNoBody(w, http.StatusBadRequest, err)
		return
	}
	response_types.OkJsonBody(w, GetStatisticsResponseBodyData{
		Posts: toResponse(posts),
	})
}

func toResponse(posts []statisticsrepo.Post) (ps []Post) {
	for _, post := range posts {
		ps = append(ps, Post(post))
	}
	return
}

type Post struct {
	Title                         string                           `json:"title"`
	PermaLinkPath                 string                           `json:"perma_link_path"`
	DataKsId                      string                           `json:"data_ks_id"`
	Score                         int32                            `json:"score"`
	SubredditId                   string                           `json:"subreddit_id"`
	CommentCount                  int32                            `json:"comment_count"`
	SubredditName                 string                           `json:"subreddit_name"`
	PolledTime                    time.Time                        `json:"polled_time"`
	AuthorId                      string                           `json:"author_id"`
	AuthorName                    string                           `json:"author_name"`
	PolledTimeRoundedMinute       time.Time                        `json:"polled_time_rounded_min"`
	Rank                          int32                            `json:"rank"`
	RankOrderType                 statisticsrepo.OrderByAlgo       `json:"rank_order_type"`
	RankOrderForCreatedWithinPast statisticsrepo.CreatedWithinPast `json:"rank_order_created_within_past"`
	Id                            int64                            `json:"id"`
}

type GetStatisticsResponseBodyData struct {
	Posts []Post `json:"posts"`
}
type GetStatisticsResponseBody = response_types.Response[GetStatisticsResponseBodyData]
type ErrorResponse = response_types.Response[struct{}]
