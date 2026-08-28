package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pb "github.com/saisai/grpc-todo/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PgxRepository struct {
	db *pgxpool.Pool
}

func NewPgxReposistory(db *pgxpool.Pool) *PgxRepository {
	return &PgxRepository{db: db}
}

func (r *PgxRepository) Create(ctx context.Context, title, description string) (*pb.Todo, error) {
	if title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	var (
		id        string
		CreatedAt time.Time
	)

	err := r.db.QueryRow(ctx, ` 
		INSERT INTO todos (title, description)
		VALUES ($1, $2)
		RETURNINT id::text, created_at
	`, title, description).Scan(&id, &CreatedAt)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create todo")
	}

	return &pb.Todo{
		Id:          id,
		Title:       title,
		Description: description,
	}, nil
}

func (r *PgxRepository) Get(ctx context.Context, id string) (*pb.Todo, error) {
	var (
		tid, title, description string
		completed               bool
		createdAt               time.Time
	)

	err := r.db.QueryRow(ctx, `
		SELECT id::text, title, description, completed, created_at
		FROM todos WHERE id = $1
	`, id).Scan(&tid, &title, &description, &completed, &createdAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "todo with id %s not found", id)
		}
		return nil, status.Error(codes.Internal, "failed to get todo")
	}

	return &pb.Todo{
		Id:          tid,
		Title:       title,
		Description: description,
		Completed:   completed,
		CreatedAt:   createdAt.Unix(),
	}, nil
}

func (r *PgxRepository) List(ctx context.Context, completed *bool) ([]*pb.Todo, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if completed != nil {
		rows, err = r.db.Query(ctx, `
			SELECT id::text, title, description, completed, created_at
			FROM todos WHERE completed = $1
			ORDER BY created_at DESC
		`, *completed)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id::text, title, description, completed, created_at
			FROM todos
			ORDER BY created_at DESC
		`)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list todos")
	}
	defer rows.Close()

	var todos []*pb.Todo
	for rows.Next() {
		var (
			id, title, description string
			comp                   bool
			createdAt              time.Time
		)
		if err := rows.Scan(&id, &title, &description, &comp, &createdAt); err != nil {
			return nil, status.Error(codes.Internal, "failed to scan todo")
		}
		todos = append(todos, &pb.Todo{
			Id:          id,
			Title:       title,
			Description: description,
			Completed:   comp,
			CreatedAt:   createdAt.Unix(),
		})
	}
	return todos, nil
}

func (r *PgxRepository) Update(ctx context.Context, id, title, description string, completed bool) (*pb.Todo, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM todos WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check todo")
	}
	if !exists {
		return nil, status.Errorf(codes.NotFound, "todo with id %s not found", id)
	}

	var (
		tid, tTitle, tDesc string
		comp               bool
		createdAt          time.Time
	)

	err = r.db.QueryRow(ctx, `
		UPDATE todos
		SET title = COALESCE(NULLIF($2, ''), title),
			description = COALESCE(NULLIF($3, ''), description),
			completed = $4
		WHERE id = $1
		RETURNING id::text, title, description, completed, created_at
	`, id, title, description, completed).Scan(&tid, &tTitle, &tDesc, &comp, &createdAt)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update todo")
	}
	return &pb.Todo{
		Id:          tid,
		Title:       tTitle,
		Description: tDesc,
		Completed:   comp,
		CreatedAt:   createdAt.Unix(),
	}, nil
}

func (r *PgxRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return status.Error(codes.Internal, "failed to delete todo")
	}
	if result.RowsAffected() == 0 {
		return status.Errorf(codes.NotFound, "todo with id %s not found", id)
	}
	return nil
}
