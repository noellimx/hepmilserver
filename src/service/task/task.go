package task

import (
	"fmt"

	"github.com/noellimx/hepmilserver/src/infrastructure/repositories/task"
)

type Service struct {
	repo *task.Repo
}

func New(repo *task.Repo) *Service {
	return &Service{repo: repo}
}

func (s Service) Create(name string, count int64, _interval string, by string, past string) error {
	orderBy := task.OrderByAlgo(by)
	interval := task.Interval(_interval)
	if name == "" || count <= 0 || interval == "" || orderBy == "" || past == "" {
		return fmt.Errorf("invalid params, name %v, count %v, interval %v, by %v, past %v", name, count, interval, orderBy, past)
	}

	if interval != task.IntervalHour {
		return fmt.Errorf("invalid params, interval requested at %v but only support %v", interval, task.IntervalHour)
	}
	return s.repo.Create(name, count, interval, orderBy, task.CreatedWithinPast(past))
}

func (s Service) Delete(subRedditName string) error {
	return s.repo.Delete(subRedditName)
}

func (s Service) GetTasks(interval task.Interval) ([]task.Task, error) {
	tasks, err := s.repo.GetByTaskInterval(interval)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}
