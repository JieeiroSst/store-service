FROM oven/bun:1-alpine AS deps

WORKDIR /app

COPY package.json bun.lockb* ./
RUN bun install --production

FROM oven/bun:1-alpine

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000

COPY --from=deps /app/node_modules ./node_modules
COPY package.json tsconfig.json server.ts ./
COPY config ./config
COPY internal ./internal

USER bun

EXPOSE 3000

ENTRYPOINT ["bun", "run", "server.ts"]
