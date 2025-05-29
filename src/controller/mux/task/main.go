package task

import (
	"encoding/json"
	"github.com/noellimx/hepmilserver/src/controller/response_types"
	"github.com/noellimx/hepmilserver/src/httplog"
	taskservice "github.com/noellimx/hepmilserver/src/service/task"
	"log"
	"net/http"
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
	ItemsCreatedWithinPast string `json:"items_created_within_past"` // ["day","hour","month","year"]
}

// Create godoc
// @Summary      Create a new task to mine subreddit periodically
// @Description  Schedule a job to get subreddit with the given parameters.
// @Tags         subreddit
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
	SubredditName string `json:"subreddit_name"`
}

// Delete godoc
// @Summary      Removes a task
// @Description  Schedule a job to get subreddit with the given parameters.
// @Tags         subreddit
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

	err := h.service.Delete(form.SubredditName)
	if err != nil {
		log.Printf("%s error=%v\n", prefix, err)
		response_types.ErrorNoBody(w, http.StatusBadRequest, err)
		return
	}
	response_types.OkJsonBody(w, struct {
	}{})
}

type ErrorResponse = response_types.Response[struct{}]
