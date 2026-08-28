package repository

import (
	"context"

	pb "github.com/saisai/grpc-todo/proto"
)

// SimpleMock is a hand-written mock useful when you don't want testify.
type SimpleMock struct {
	CreateFunc func(ctx context.Context, title, description string) (*pb.Todo, error)
	GetFunc func(ctx context.Context, id string) (*pb.Todo, error)
	ListFunc func(ctx context.Context, completed *bool) ([]*pb.Todo, error)
	UpdateFunc func(ctx context.Context, id, title, description string, completed bool) (*pb.Todo, error)
	DeleteFunc func(ctx context.Context, id string) error 
}

func (m *SimpleMock) Create(ctx context.Context, title, description string) (*pb.Todo, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, title, description)
	}
	return nil, nil 
}

func (m *SimpleMock) Get(ctx context.Context, id string) (*pb.Todo, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil 
}

func (m *SimpleMock) List(ctx context.Context, completed *bool) ([]*pb.Todo, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, completed)
	}
	return nil, nil 
}

func (m *SimpleMock) Update(ctx context.Context, id, title, description string, completed bool) (*pb.Todo, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, title, description, completed)
	}
	return nil, nil 
}

func (m *SimpleMock) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil 
}
