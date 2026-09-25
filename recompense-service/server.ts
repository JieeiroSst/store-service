import config from "./config";
import { Container } from "./internal/bootstrap/container";

const container = await Container.create(config);

const server = Bun.serve({
  port: config.Port,
  fetch: (request) => container.handler.handle(request),
});

console.log(`Listening on localhost:${server.port}`);

for (const signal of ["SIGINT", "SIGTERM"] as const) {
  process.on(signal, async () => {
    await server.stop();
    await container.close();
    process.exit(0);
  });
}
