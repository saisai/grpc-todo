package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
	"uuid"

	pb "github.com/saisai/grpc-todo/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const port = ":50051"

// server implements the generated TodoServiceServer interface
type server struct {
	pb.UnimplementedTodoServiceServer // required for forward compatibility
	mu                                sync.RWMutex
	todos                             map[string]*pb.Todo // in-memory storage
}

func newServer() *server {
	return &server{
		todos: make(map[string]*pb.Todo),
	}
}

func (s *server) CreateTodo(ctx context.Context, req *pb.CreateTodoRequest) (*pb.Todo, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	todo := &pb.Todo{
		Id:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   time.Now().Unix(),
	}

	s.mu.Lock()
	s.todos[todo.Id] = todo
	s.mu.Unlock()

	log.Printf("Created todo: %s - %s", todo.Id, todo.Title)
	return todo, nil
}

func (s *server) GetTodo(ctx context.Context, req *pb.GetTodoRequest) (*pb.Todo, error) {
	s.mu.RLock()
	todo, ok := s.todos[req.Id]
	s.mu.RUnlock()

	if !ok {
		return nil, status.Errorf(codes.NotFound, "todo with id %s not found", req.Id)
	}
	return todo, nil
}

func (s *server) ListTodos(ctx context.Context, req *pb.ListTodosRequest) (*pb.ListTodosResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var todos []*pb.Todo
	for _, t := range s.todos {
		// Applly optional filter
		if req.Completed != nil && t.Completed != *req.Completed {
			continue
		}
		todos = append(todos, t)
	}
	return &pb.ListTodosResponse{Todos: todos}, nil
}

func (s *server) UpdateTodo(ctx context.Context, req *pb.UpdateTodoRequest) (*pb.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, ok := s.todos[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "todo with id %s not found", req.Id)
	}

	if req.Title != "" {
		todo.Title = req.Title
	}
	if req.Description != "" {
		todo.Description = req.Description
	}
	todo.Completed = req.Completed

	log.Printf("Updated todo: %d", todo.Id)
	return todo, nil
}

func (s *server) DeleteTodo(ctx context.Context, req *pb.DeleteTodoRequest) (*pb.DeleteTodoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.todos[req.Id]; !ok {
		return nil, status.Errorf(codes.NotFound, "todo with id %s not found", req.Id)
	}

	delete(s.todos, req.Id)
	log.Printf("Delete todo: %s", req.Id)
	return &pb.DeleteTodoResponse{Success: true}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTodoServiceServer(s, newServer())

	fmt.Printf("🚀 gRPC Todo Server listening on %s\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
