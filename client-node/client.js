const grpc = require("@grpc/grpc-js");
const protoLoader = require("@grpc/proto-loader");
const path = require("path");

const PROTO_PATH = path.join(__dirname, "../proto/todo.proto");

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

const todoProto = grpc.loadPackageDefinition(packageDefinition).todo;

function main() {
  const client = new todoProto.TodoService(
    "localhost:50051",
    grpc.credentials.createInsecure()
  );

  // Helper to turn callback into Promise
  const call = (method, request) =>
    new Promise((resolve, reject) => {
      client[method](request, (err, response) => {
        if (err) reject(err);
        else resolve(response);
      });
    });

  (async () => {
    try {
      console.log("=== Creating Todos ===");
      const todo1 = await call("CreateTodo", {
        title: "Learn gRPC with Node.js",
        description: "Implement the Node.js client",
      });
      console.log(`Created: [${todo1.id}] ${todo1.title}`);

      const todo2 = await call("CreateTodo", {
        title: "Ship the tutorial",
        description: "Make sure everything works end-to-end",
      });
      console.log(`Created: [${todo2.id}] ${todo2.title}`);

      console.log("\n=== List All Todos ===");
      let list = await call("ListTodos", {});
      list.todos.forEach((t) => {
        console.log(`- [${t.id}] ${t.title} (completed: ${t.completed})`);
      });

      console.log("\n=== Get Todo ===");
      const got = await call("GetTodo", { id: todo1.id });
      console.log(`Got: ${got.title} - ${got.description}`);

      console.log("\n=== Update Todo ===");
      const updated = await call("UpdateTodo", {
        id: todo1.id,
        title: "Learn gRPC with Node.js (done)",
        description: "Node client works!",
        completed: true,
      });
      console.log(`Updated: ${updated.title} (completed: ${updated.completed})`);

      console.log("\n=== List Completed Todos ===");
      list = await call("ListTodos", { completed: true });
      list.todos.forEach((t) => {
        console.log(`- [${t.id}] ${t.title}`);
      });

      console.log("\n=== Delete Todo ===");
      const del = await call("DeleteTodo", { id: todo2.id });
      console.log(`Deleted successfully: ${del.success}`);

      console.log("\n✅ Node.js client demo finished!");
    } catch (err) {
      console.error("Error:", err.message);
      process.exit(1);
    }
  })();
}

main();