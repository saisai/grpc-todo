package main

import (
	"context"
	"testing"
	"time"

	pb "github.com/saisai/grpc-todo/proto"
	"github.com/saisai/grpc-todo/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateTOdo_Success(t *testing.T) {
	mockRepo := new(repository.MockTodoRepository)
	svc := newServer(mockRepo)

	expected := &pb.Todo{
		Id:          "abc-123",
		Title:       "Learn gRPC",
		Description: "Write tests",
		Completed:   false,
		CreatedAt:   time.Now().Unix(),
	}

	mockRepo.On("Create", mock.Anything, "Learn gRPC", "Write tests").
		Return(expected, nil)

	resp, err := svc.CreateTodo(context.Background(), &pb.CreateTodoRequest{
		Title:       "Learn gRPC",
		Description: "Write tests",
	})

	require.NoError(t, err)
	assert.Equal(t, expected.Id, resp.Id)
	assert.Equal(t, "Learn gRPC", resp.Title)
	mockRepo.AssertExpectations(t)
}

func TestCreateTodo_EmptyTitle(t *testing.T) {
	mockRepo := new(repository.MockTodoRepository)
	svc := newServer(mockRepo)

	// The repository itself returns the error
	mockRepo.On("Create", mock.Anything, "", "desc").
		Return(nil, status.Error(codes.InvalidArgument, "title is required"))

	_, err := svc.CreateTodo(context.Background(), &pb.CreateTodoRequest{
		Title:       "",
		Description: "desc",
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	mockRepo.AssertExpectations(t)
}

func TestGetTodo_Success(t *testing.T) {
	mockRepo := new(repository.MockTodoRepository)
	svc := newServer(mockRepo)

	expected := &pb.Todo{
		Id:          "abc-123",
		Title:       "Existing todo",
		Description: "Hello",
		Completed:   false,
	}

	mockRepo.On("Get", mock.Anything, "abc-123").Return(expected, nil)

	resp, err := svc.GetTodo(context.Background(), &pb.GetTodoRequest{Id: "abc-123"})

	require.NoError(t, err)
	assert.Equal(t, "abc-123", resp.Id)
	assert.Equal(t, "Existing todo", resp.Title)
	mockRepo.AssertExpectations(t)
}

func TestGetTodo_NotFound(t *testing.T) {
	mockRepo := new(repository.MockTodoRepository)
	svc := newServer(mockRepo)

	mockRepo.On("Get", mock.Anything, "missing-id").
		Return(nil, status.Error(codes.NotFound, "todo with id missing-id not found"))

	_, err := svc.GetTodo(context.Background(), &pb.GetTodoRequest{Id: "missing-id"})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	mockRepo.AssertExpectations(t)
}

func TestListTodos_All(t *testing.T) {
	mockRepo := new(repository.MockTodoRepository)
	svc := newServer(mockRepo)

	todos := []*pb.Todo{
		{Id: "1", Title: "Todo 1"},
		{Id: "2", Title: "Todo 2"},
	}

	mockRepo.On("List", mock.Anything, (*bool)(nil)).Return(todos, nil)

	resp, err := svc.ListTodos(context.Background(), &pb.ListTodosRequest{})

	require.NoError(t, err)
	assert.Len(t, resp.Todos, 2)
	mockRepo.AssertExpectations(t)
}

func TestListTodos_OnlyCompleted(t *testing.T) {
	mockRepo := new(repository.MockTodoRepository)
	svc := newServer(mockRepo)

	completed := true
	todos := []*pb.Todo{
		{Id: "1", Title: "Done", Completed: true},
	}

	mockRepo.On("List", mock.Anything, &completed).Return(todos, nil)

	resp, err := svc.ListTodos(context.Background(), &pb.ListTodosRequest{
		Completed: &completed,
	})

	require.NoError(t, err)
	assert.Len(t, resp.Todos, 1)
	assert.True(t, resp.Todos[0].Completed)
	mockRepo.AssertExpectations(t)
}

func TestUpdateTodo_Success(t *testing.T) {
	mockRepo := new(repository.MockTodoRepository)
	svc := newServer(mockRepo)

	updated := &pb.Todo{
		Id:          "abc-123",
		Title:       "Updated title",
		Description: "Updated desc",
		Completed:   true,
	}

	mockRepo.On("Update", mock.Anything, "abc-123", "Updated title", "Updated desc", true).
		Return(updated, nil)

	resp, err := svc.UpdateTodo(context.Background(), &pb.UpdateTodoRequest{
		Id:          "abc-123",
		Title:       "Updated title",
		Description: "Updated desc",
		Completed:   true,
	})

	require.NoError(t, err)
	assert.Equal(t, "Updated title", resp.Title)
	assert.True(t, resp.Completed)
	mockRepo.AssertExpectations(t)
}

func TestDeleteTodo_Success(t *testing.T) {
	mockRepo := new(repository.MockTodoRepository)
	svc := newServer(mockRepo)

	mockRepo.On("Delete", mock.Anything, "abc-123").Return(nil)

	resp, err := svc.DeleteTodo(context.Background(), &pb.DeleteTodoRequest{Id: "abc-123"})

	require.NoError(t, err)
	assert.True(t, resp.Success)
	mockRepo.AssertExpectations(t)
}

func TestDeleteTodo_NotFound(t *testing.T) {
	mockRepo := new(repository.MockTodoRepository)
	svc := newServer(mockRepo)

	mockRepo.On("Delete", mock.Anything, "missing").
		Return(status.Error(codes.NotFound, "todo with id missing not found"))

	_, err := svc.DeleteTodo(context.Background(), &pb.DeleteTodoRequest{Id: "missing"})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	mockRepo.AssertExpectations(t)
}
