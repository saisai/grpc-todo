package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/saisai/grpc-todo/proto"
	"github.com/saisai/grpc-todo/server/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

const port = ":50051"

// server implements the generated TodoServiceServer interface
type server struct {
	pb.UnimplementedTodoServiceServer // required for forward compatibility
	repo                              repository.TodoRepository
}

func newServer(repo repository.TodoRepository) *server {
	return &server{repo: repo}
}

func (s *server) CreateTodo(ctx context.Context, req *pb.CreateTodoRequest) (*pb.Todo, error) {
	return s.repo.Create(ctx, req.Title, req.Description)
}

func (s *server) GetTodo(ctx context.Context, req *pb.GetTodoRequest) (*pb.Todo, error) {
	return s.repo.Get(ctx, req.Id)
}

func (s *server) ListTodos(ctx context.Context, req *pb.ListTodosRequest) (*pb.ListTodosResponse, error) {
	todos, err := s.repo.List(ctx, req.Completed)
	if err != nil {
		return nil, err
	}
	return &pb.ListTodosResponse{Todos: todos}, nil
}

func (s *server) UpdateTodo(ctx context.Context, req *pb.UpdateTodoRequest) (*pb.Todo, error) {
	return s.repo.Update(ctx, req.Id, req.Title, req.Description, req.Completed)
}

func (s *server) DeleteTodo(ctx context.Context, req *pb.DeleteTodoRequest) (*pb.DeleteTodoResponse, error) {
	if err := s.repo.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &pb.DeleteTodoResponse{Success: true}, nil
}

func main() {
	// ============================================================
	// SWITCH DRIVER HERE
	// ============================================================
	// Options: "memory" | "pgx" | "pq"
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "memory" // default for easy local testing
	}

	var repo repository.TodoRepository

	switch driver {
	case "memory":
		repo = repository.NewMemoryRepository()
		log.Println("✅ Using in-memory repository")
	case "pgx":
		dsn := getDSN()
		pool, err := pgxpool.New(context.Background(), dsn)
		if err != nil {
			log.Fatalf("pgx: unable to connect: %v", err)
		}
		if err := pool.Ping(context.Background()); err != nil {
			log.Fatalf("pgx: ping failed: %v", err)
		}
		repo = repository.NewPgxReposistory(pool)
		log.Println("✅ Using PostgreSQL with pgx")

	case "pq":
		dsn := getDSN()
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			log.Fatalf("pq: unable to open: %v", err)
		}
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		if err := db.Ping(); err != nil {
			log.Fatalf("pq: ping failed: %v", err)
		}
		if err := ensureSchemaSQL(db); err != nil {
			log.Fatalf("pq: schema failed: %v", err)
		}
		repo = repository.NewPqRepository(db)
		log.Println("✅ Using PostgreSQL with lib/pq")

	default:
		log.Fatalf("unknown DB_DRIVER: %s (use memory|pgx|pq)", driver)
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTodoServiceServer(s, newServer(repo))

	fmt.Printf("🚀 gRPC Todo Server listening on %s\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func getDSN() string {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable"
	}
	return dsn
}

func ensureSchemaPgx(pool *pgxpool.Pool) error {
	_, err := pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS todos (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title       TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			completed   BOOLEAN NOT NULL DEFAULT FALSE,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
	`)
	return err
}

func ensureSchemaSQL(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title       TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			completed   BOOLEAN NOT NULL DEFAULT FALSE,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
	`)
	return err
}
