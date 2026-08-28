package repository

import (
	"context"
	"sync"
	"time"
	"uuid"

	pb "github.com/saisai/grpc-todo/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MemoryReposiory struct {
	mu    sync.RWMutex
	todos map[string]*pb.Todo
}

func NewMemoryRepository() *MemoryReposiory {
	return &MemoryReposiory{
		todos: make(map[string]*pb.Todo),
	}
}

func (r *MemoryReposiory) Create(ctx context.Context, title, description string) (*pb.Todo, error) {
	if title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	todo := &pb.Todo{
		Id:          uuid.New().String(),
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now().Unix(),
	}

	r.mu.Lock()
	r.todos[todo.Id] = todo
	r.mu.Unlock()

	return todo, nil
}

func (r *MemoryReposiory) Get(ctx context.Context, id string) (*pb.Todo, error) {
	r.mu.RLock()
	todo, ok := r.todos[id]
	r.mu.RUnlock()

	if !ok {
		return nil, status.Errorf(codes.NotFound, "todo with id %s not found", id)
	}
	return todo, nil
}

func (r *MemoryReposiory) List(ctx context.Context, completed *bool) ([]*pb.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*pb.Todo
	for _, t := range r.todos {
		if completed != nil && t.Completed != *completed {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

func (r *MemoryReposiory) Update(ctx context.Context, id, title, description string, completed bool) (*pb.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	todo, ok := r.todos[id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "todo with id %s not found", id)
	}

	if title != "" {
		todo.Title = title
	}
	if description != "" {
		todo.Description = description
	}
	todo.Completed = completed

	return todo, nil
}

func (r *MemoryReposiory) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.todos[id]; !ok {
		return status.Errorf(codes.NotFound, "todo with id %s not found", id)
	}

	delete(r.todos, id)
	return nil
}
