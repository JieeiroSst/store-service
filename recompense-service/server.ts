const server = Bun.serve({
  port: 3000 || process.env.PORT,
  fetch(request) {
    return new Response("Welcome to Bun!");
  },
});

console.log(`Listening on localhost:${server.port}`);