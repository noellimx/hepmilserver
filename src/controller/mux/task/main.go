package task

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/noellimx/hepmilserver/src/controller/response_types"
	"github.com/noellimx/hepmilserver/src/httplog"
	taskrepo "github.com/noellimx/hepmilserver/src/infrastructure/repositories/task"
	taskservice "github.com/noellimx/hepmilserver/src/service/task"
)

type Handlers struct {
	service *taskservice.Service
}

func NewHandlers(service *taskservice.Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

type CreateRequestBody struct {
	SubredditName          string `json:"subreddit_name"`            // Subreddit Name
	MinItemCount           int64  `json:"min_item_count"`            // Minimum Item Count to retrieve
	Interval               string `json:"interval"`                  // to be executed every interval ["hour"]
	OrderBy                string `json:"order_by"`                  // ["top", "hot", "best", "new"]
	ItemsCreatedWithinPast string `json:"posts_created_within_past"` // ["day","hour","month","year"]
}

// Create godoc
// @Summary      Create a new task to mine subreddit periodically
// @Description  Schedule a job to get subreddit with the given parameters.
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        request body CreateRequestBody true "Create Request Body"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  ErrorResponse
// @Router       /task [post]
func (h Handlers) Create(w http.ResponseWriter, r *http.Request) {
	prefix := httplog.SPrintHttpRequestPrefix(r)
	form := &CreateRequestBody{}
	json.NewDecoder(r.Body).Decode(form)

	err := h.service.Create(form.SubredditName, form.MinItemCount, form.Interval, form.OrderBy, form.ItemsCreatedWithinPast)
	if err != nil {
		log.Printf("%s error=%v\n", prefix, err)
		response_types.ErrorNoBody(w, http.StatusBadRequest, err)
		return
	}
	response_types.OkJsonBody(w, struct {
	}{})
}

type DeleteRequestBody struct {
	Id int64 `json:"id"`
}

// Delete godoc
// @Summary      Removes a task
// @Description  Schedule a job to get subreddit with the given parameters.
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        request body DeleteRequestBody true "Delete Request Body"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  ErrorResponse
// @Router       /task [delete]
func (h Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	prefix := httplog.SPrintHttpRequestPrefix(r)
	form := &DeleteRequestBody{}
	json.NewDecoder(r.Body).Decode(form)

	err := h.service.Delete(form.Id)
	if err != nil {
		log.Printf("%s error=%v\n", prefix, err)
		response_types.ErrorNoBody(w, http.StatusBadRequest, err)
		return
	}
	response_types.OkJsonBody(w, struct {
	}{})
}

// List godoc
// @Summary      Get tasks.
// @Description  Get tasks.
// @Tags         task
// @Accept       json
// @Produce      json
// @Success      200  {object}  ListResponseBodyData
// @Failure      500  {object}  ErrorResponse
// @Router       /tasks [delete]
func (h Handlers) List(w http.ResponseWriter, r *http.Request) {
	prefix := httplog.SPrintHttpRequestPrefix(r)
	tasks, err := h.service.GetTasks()
	if err != nil {
		log.Printf("%s error=%v\n", prefix, err)
		response_types.ErrorNoBody(w, http.StatusBadRequest, err)
		return
	}
	response_types.OkJsonBody(w, ListResponseBodyData{
		Tasks: toResponse(tasks),
	})
}

func toResponse(tasks []taskrepo.Task) (tt []Task) {
	for _, t := range tasks {
		tt = append(tt, Task{
			Id:                     t.Id,
			SubRedditName:          t.SubRedditName,
			MinItemCount:           t.MinItemCount,
			Interval:               Granularity(t.Interval),
			OrderBy:                OrderByAlgo(t.OrderBy),
			PostsCreatedWithinPast: CreatedWithinPast(t.PostsCreatedWithinPast),
		})
	}
	return
}

type OrderByAlgo string

const (
	OrderByAlgoTop  OrderByAlgo = "top"
	OrderByAlgoBest OrderByAlgo = "best"
	OrderByAlgoHot  OrderByAlgo = "hot"
	OrderByAlgoNew  OrderByAlgo = "new"
)

type Granularity string

const (
	GranularityHour Granularity = "hour"
)

type CreatedWithinPast string

const (
	CreatedWithinPastHour  CreatedWithinPast = "hour"
	CreatedWithinPastDay   CreatedWithinPast = "day"
	CreatedWithinPastMonth CreatedWithinPast = "month"
	CreatedWithinPastYear  CreatedWithinPast = "year"
)

type Task struct {
	Id                     int64             `json:"id"`
	SubRedditName          string            `json:"subreddit_name"`
	MinItemCount           int64             `json:"min_item_count"`
	Interval               Granularity       `json:"interval"`
	OrderBy                OrderByAlgo       `json:"order_by"`
	PostsCreatedWithinPast CreatedWithinPast `json:"posts_created_within_past"`
}

type ListResponseBodyData struct {
	Tasks []Task `json:"tasks"`
}

type ErrorResponse = response_types.Response[struct{}]
