package repository

import (
	"context"

	pb "github.com/saisai/grpc-todo/proto"
	"github.com/stretchr/testify/mock"
)

// MockTodoRepository is a testify/mock implementation of TodoRepository.
// Use it in unit tests so you don't need a real database.
type MockTodoRepository struct {
	mock.Mock
}

func (m *MockTodoRepository) Create(ctx context.Context, title, description string) (*pb.Todo, error) {
	args := m.Called(ctx, title, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Todo), args.Error(1)
}

func (m *MockTodoRepository) Get(ctx context.Context, id string) (*pb.Todo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Todo), args.Error(1)
}

func (m *MockTodoRepository) List(ctx context.Context, completed *bool) ([]*pb.Todo, error) {
	args := m.Called(ctx, completed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pb.Todo), args.Error(1)
}

func (m *MockTodoRepository) Update(ctx context.Context, id, title, description string, completed bool) (*pb.Todo, error) {
	args := m.Called(ctx, id, title, description, completed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Todo), args.Error(1)
}

func (m *MockTodoRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
