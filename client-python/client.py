import grpc 
import todo_pb2
import todo_pb2_grpc

def run():
    with grpc.insecure_channel("localhost:50051") as channel:
        stub = todo_pb2_grpc.TodoServiceStub(channel)

        # Create
        todo1 = stub.CreateTodo(todo_pb2.CreateTodoRequest(
            title="Learn gRPC with Python",
            description="Implement the Python client",
        ))

        todo2 = stub.CreateTodo(todo_pb2.CreateTodoRequest(
            title="Test Node.js client",
            description="Verify interop with Node",
        ))
        print(f"Created: [{todo2.id}] {todo2.title}")

        # List
        print("\n=== List All Todos ===")
        response = stub.ListTodos(todo_pb2.ListTodosRequest())
        for t in response.todos:
            print(f"- [{t.id}] {t.title} (completed: {t.completed})")

        # Get
        print("\n=== Get Todo ===")
        got = stub.GetTodo(todo_pb2.GetTodoRequest(id=todo1.id))
        print(f"Got: {got.title} - {got.description}")

        # Update
        print("\n=== Update Todo ===")
        updated = stub.UpdateTodo(todo_pb2.UpdateTodoRequest(
            id=todo1.id,
            title="Learn gRPC with Python (done)",
            description="Python client works!",
            completed=True,
        ))
        print(f"Updated: {updated.title} (completed: {updated.completed})")

        # List completed
        print("\n=== List Completed Todos ===")
        response = stub.ListTodos(todo_pb2.ListTodosRequest(completed=True))
        for t in response.todos:
            print(f"- [{t.id}] {t.title}")

        # Delete
        print("\n=== Delete Todo ===")
        del_resp = stub.DeleteTodo(todo_pb2.DeleteTodoRequest(id=todo2.id))
        print(f"Deleted successfully: {del_resp.success}")

        print("\n✅ Python client demo finished!")

if __name__ == "__main__":
    run()