package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/saisai/grpc-todo/proto"
)

func main() {
	// 1. Connect to the server
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	defer conn.Close()

	client := pb.NewTodoServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Create
	fmt.Println("=== Creating Todos ===")
	todo1, err := client.CreateTodo(ctx, &pb.CreateTodoRequest{
		Title:       "Learn gRPC",
		Description: "Complete the gRPC todo tutorial",
	})

	if err != nil {
		log.Fatalf("CreateTodo failed: %v", err)
	}
	fmt.Printf("Created: [%s] %s\n", todo1.Id, todo1.Title)

	todo2, _ := client.CreateTodo(ctx, &pb.CreateTodoRequest{
		Title:       "Write Python client",
		Description: "Implement the Python gRPC client",
	})
	fmt.Printf("Created: [%s] %s\n", todo2.Id, todo2.Title)

	// 3. List
	fmt.Println("\n=== List All Todos ===")
	listResp, _ := client.ListTodos(ctx, &pb.ListTodosRequest{})
	for _, t := range listResp.Todos {
		fmt.Printf("- [%s] %s (completed: %v)\n", t.Id, t.Title, t.Completed)
	}

	// 4. Get
	fmt.Println("\n=== Get Todo ===")
	got, _ := client.GetTodo(ctx, &pb.GetTodoRequest{Id: todo1.Id})
	fmt.Printf("Got: %s - %s\n", got.Title, got.Description)

	// 5. Update
	fmt.Println("\n=== Update Todo ===")
	updated, _ := client.UpdateTodo(ctx, &pb.UpdateTodoRequest{
		Id:          todo1.Id,
		Title:       "Learn gRPC (done)",
		Description: "Finished the tutorial",
		Completed:   true,
	})
	fmt.Printf("Updated: %s (completed: %v)\n", updated.Title, updated.Completed)

	// 6. List only completed
	fmt.Println("\n=== List Completed Todos ===")
	completed := true
	listResp, _ = client.ListTodos(ctx, &pb.ListTodosRequest{Completed: &completed})
	for _, t := range listResp.Todos {
		fmt.Printf("- [%s] %s\n", t.Id, t.Title)
	}

	// 7. Delete
	fmt.Println("\n=== Delete Todo ===")
	delResp, _ := client.DeleteTodo(ctx, &pb.DeleteTodoRequest{Id: todo2.Id})
	fmt.Printf("Deleted successfully: %v\n", delResp.Success)

	fmt.Println("\n✅ Go client demo finished!")

}
