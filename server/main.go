package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/saisai/grpc-todo/proto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const port = ":50051"

// server implements the generated TodoServiceServer interface
type server struct {
	pb.UnimplementedTodoServiceServer // required for forward compatibility
	db                                *pgxpool.Pool
}

func newServer(db *pgxpool.Pool) *server {
	return &server{db: db}
}

func (s *server) CreateTodo(ctx context.Context, req *pb.CreateTodoRequest) (*pb.Todo, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	var (
		id        string
		createdAt time.Time
	)

	err := s.db.QueryRow(ctx, ` 
		INSERT INTO todos (title, description)
		VALUES ($1, $2)
		RETURNING id::text, created_at
	`, req.Title, req.Description).Scan(&id, &createdAt)

	if err != nil {
		log.Printf("CreateTodo error: %v", err)
		return nil, status.Error(codes.Internal, "failed to create todo")
	}

	todo := &pb.Todo{
		Id:          id,
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   createdAt.Unix(),
	}

	log.Printf("Created todo: %s - %s", todo.Id, todo.Title)
	return todo, nil
}

func (s *server) GetTodo(ctx context.Context, req *pb.GetTodoRequest) (*pb.Todo, error) {
	var (
		id, title, description string
		completed              bool
		createdAt              time.Time
	)

	err := s.db.QueryRow(ctx, `
		SELECT id::text, title, description, completed, created_at
		FROM todos
		WHERE id = $1
	`, req.Id).Scan(&id, &title, &description, &completed, &createdAt)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, status.Errorf(codes.NotFound, "todo with id %s not found", req.Id)
		}
		log.Printf("GetTodo errorr: %v", err)
		return nil, status.Error(codes.Internal, "failed to get todo")
	}

	return &pb.Todo{
		Id:          id,
		Title:       title,
		Description: description,
		Completed:   completed,
		CreatedAt:   createdAt.Unix(),
	}, nil
}

func (s *server) ListTodos(ctx context.Context, req *pb.ListTodosRequest) (*pb.ListTodosResponse, error) {
	var rows pgx.Rows
	var err error

	if req.Completed != nil {
		rows, err = s.db.Query(ctx, `
			SELECT id::text, title, description, completed, created_at
			FROM todos
			WHERE completed = $1
			ORDER BY created_at DESC
		`, *req.Completed)
	} else {
		rows, err = s.db.Query(ctx, `
			SELECT id::text, title, description, completed, created_at
			FROM todos
			ORDER BY created_at DESC
		`)
	}

	if err != nil {
		log.Printf("ListTodos error: %v", err)
		return nil, status.Error(codes.Internal, "failed to list todos")
	}
	defer rows.Close()

	var todos []*pb.Todo
	for rows.Next() {
		var (
			id, title, description string
			completed              bool
			createdAt              time.Time
		)
		if err := rows.Scan(&id, &title, &description, &completed, &createdAt); err != nil {
			return nil, status.Error(codes.Internal, "failed to scan todo")
		}
		todos = append(todos, &pb.Todo{
			Id:          id,
			Title:       title,
			Description: description,
			Completed:   completed,
			CreatedAt:   createdAt.Unix(),
		})
	}
	return &pb.ListTodosResponse{Todos: todos}, nil
}

func (s *server) UpdateTodo(ctx context.Context, req *pb.UpdateTodoRequest) (*pb.Todo, error) {
	// First check if the todo exists
	var exists bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM todos WHERE id = $1)`, req.Id).Scan(&exists)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check todo")
	}
	if !exists {
		return nil, status.Errorf(codes.NotFound, "todo with id %s not found", req.Id)
	}
	// Build dynamic update (only update provided fields)
	query := `
		UPDATE todos
		SET title = COALESCE(NULLIF($2, ''), title),
			description = COALESCE(NULLIF($3, ''), description),
			completed = $4
		WHERE id = $1
		RETURNING id::text, title, description, completed, created_at
	`
	var (
		id, title, description string
		completed              bool
		createdAt              time.Time
	)

	err = s.db.QueryRow(ctx, query, req.Id, req.Title, req.Description, req.Completed).
		Scan(&id, &title, &description, &completed, &createdAt)

	if err != nil {
		log.Printf("UpdateTodo error: %v", err)
		return nil, status.Error(codes.Internal, "failed to update todo")
	}

	log.Printf("Upated todo: %s", id)
	return &pb.Todo{
		Id:          id,
		Title:       title,
		Description: description,
		Completed:   completed,
		CreatedAt:   createdAt.Unix(),
	}, nil
}

func (s *server) DeleteTodo(ctx context.Context, req *pb.DeleteTodoRequest) (*pb.DeleteTodoResponse, error) {
	result, err := s.db.Exec(ctx, `DELETE FROM todos WHERE id = $1`, req.Id)
	if err != nil {
		log.Printf("DeleteTodo error: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete todo")
	}

	if result.RowsAffected() == 0 {
		return nil, status.Errorf(codes.NotFound, "todo with id %s not found", req.Id)
	}

	log.Printf("Deleted todo: %s", req.Id)
	return &pb.DeleteTodoResponse{Success: true}, nil
}

func main() {
	// Connection string - you can also use environment variable
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Default for lcoal development
		dsn = "postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable"
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("unable to ping database: %v", err)
	}
	log.Println(" Connected to PostgreSQL")

	// Create the table if it doesn't exist
	schema, err := os.ReadFile("server/schema.sql")
	if err != nil {
		// fallback - embed the system
		schema = []byte(`
			CREATE TABLE IF NOT EXISTS todos (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				title TEXT NOT NULL, 
				description TEXT NOT NULL DEFAULT '',
				completed BOOLEAN NOT NULL DEFAULT FALSE,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);
			CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
		`)
	}

	if _, err := db.Exec(ctx, string(schema)); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}
	log.Println(" Schema ready")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTodoServiceServer(s, newServer(db))

	fmt.Printf("🚀 gRPC Todo Server listening on %s\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
