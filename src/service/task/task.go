package task

import (
	"fmt"

	"github.com/noellimx/redditminer/src/infrastructure/repositories/task"
)

type Service struct {
	repo *task.Repo
}

func New(repo *task.Repo) *Service {
	return &Service{repo: repo}
}

func (s Service) Create(name string, count int64, _interval string, by string, past string) error {
	orderBy := task.OrderByAlgo(by)
	interval := task.Granularity(_interval)
	if name == "" || count <= 0 || interval == "" || orderBy == "" || past == "" {
		return fmt.Errorf("some invalid params, name=%v, count=%v, interval=%v, by=%v, past %v", name, count, interval, orderBy, past)
	}

	if interval != task.GranularityHour {
		return fmt.Errorf("invalid params, interval requested at %v but only support %v", interval, task.GranularityHour)
	}
	return s.repo.Create(name, count, interval, orderBy, task.CreatedWithinPast(past))
}

func (s Service) Delete(id int64) error {
	return s.repo.Delete(id)
}

func (s Service) GetTasksByInterval(interval task.Granularity) ([]task.Task, error) {
	tasks, err := s.repo.GetTasksByInterval(interval)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s Service) GetTasks() ([]task.Task, error) {
	tasks, err := s.repo.GetTasks()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}
