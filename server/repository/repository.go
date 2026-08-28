package repository

import (
	"context"

	pb "github.com/saisai/grpc-todo/proto"
)

// TodoRepository defines the storage operations for todos.
// Any concrete implementation (memory, pgx, pq, etc.) must satisfy this interface.
type TodoRepository interface {
	Create(ctx context.Context, title, description string) (*pb.Todo, error)
	Get(ctx context.Context, id string) (*pb.Todo, error)
	List(ctx context.Context, completed *bool) ([]*pb.Todo, error)
	Update(ctx context.Context, id, title, description string, completed bool) (*pb.Todo, error)
	Delete(ctx context.Context, id string) error
}
