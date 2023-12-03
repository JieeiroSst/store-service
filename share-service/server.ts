import Bao from "baojs";
import dotenv from "dotenv";
import OpenAI from "openai";

dotenv.config();

const app = new Bao();
const port = parseInt(process.env.PORT || "8080");
const openai = new OpenAI({
  apiKey: process.env.API_KEY,
});

app.get("/", (ctx) => {
  return ctx.sendText("Hello world from Bao.js running on Railway!");
});

app.post("/", async (ctx) => {
  const content = ctx.params.content
  const result = await openai.chat.completions.create({
    model: "gpt-3.5-turbo",
    messages: [{ role: 'user', content: content }],
    stream: true,
  });

  return ctx.sendJson({ result })
});


const server = app.listen({ port: port });
console.log(`Server listening on ${server.hostname}:${port}`);