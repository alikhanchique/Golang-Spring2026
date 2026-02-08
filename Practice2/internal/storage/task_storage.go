package storage

import (
	"errors"
	"sync"
	"kbtu.practice2.alikh/internal/models"
)

var ErrNotFound = errors.New("task not found")

type TaskStore struct {
	mu     sync.Mutex
	tasks  map[int]models.Task
	nextID int
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks:  make(map[int]models.Task),
		nextID: 1,
	}
}

func (s *TaskStore) GetAll() []models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]models.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}

func (s *TaskStore) GetByID(id int) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return models.Task{}, ErrNotFound
	}
	return t, nil
}

func (s *TaskStore) Create(title string) models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := models.Task{
		ID:    s.nextID,
		Title: title,
		Done:  false,
	}
	s.tasks[s.nextID] = task
	s.nextID++
	return task
}

func (s *TaskStore) UpdateDone(id int, done bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return ErrNotFound
	}

	task.Done = done
	s.tasks[id] = task
	return nil
}
