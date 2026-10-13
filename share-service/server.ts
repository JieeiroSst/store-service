import Bao from "baojs";
import dotenv from "dotenv";
import OpenAI from "openai";

dotenv.config();

const app = new Bao();
const port = parseInt(process.env.PORT || "8080");
const model = process.env.OPENAI_MODEL || "gpt-3.5-turbo";
const maxContentLength = 8000;
const openai = new OpenAI({
  apiKey: process.env.API_KEY,
  baseURL: process.env.OPENAI_BASE_URL || undefined,
});

const status = (code: number) => ({ status: code, headers: {} });

app.get("/health", (ctx) => ctx.sendJson({ status: "ok" }));

app.post("/", async (ctx) => {
  let body: unknown;
  try {
    body = await ctx.req.json();
  } catch {
    return ctx.sendJson({ error: "request body must be JSON" }, status(400));
  }

  const content = (body as { content?: unknown })?.content;
  if (typeof content !== "string" || content.trim() === "") {
    return ctx.sendJson({ error: "content is required" }, status(400));
  }
  if (content.length > maxContentLength) {
    return ctx.sendJson({ error: `content must be at most ${maxContentLength} characters` }, status(400));
  }

  try {
    const completion = await openai.chat.completions.create({
      model,
      messages: [{ role: "user", content }],
    });
    return ctx.sendJson({
      result: completion.choices[0]?.message?.content ?? "",
      model: completion.model,
    });
  } catch (err) {
    console.error("openai request failed", err);
    return ctx.sendJson({ error: "AI service is unavailable" }, status(502));
  }
});

const server = app.listen({ port: port });
console.log(`Server listening on ${server.hostname}:${port}`);
