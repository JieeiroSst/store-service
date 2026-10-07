FROM oven/bun:1-alpine AS deps

WORKDIR /app

COPY package.json bun.lockb* ./
RUN bun install --production

FROM oven/bun:1-alpine

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=8080

COPY --from=deps /app/node_modules ./node_modules
COPY package.json tsconfig.json server.ts ./

USER bun

EXPOSE 8080

ENTRYPOINT ["bun", "run", "server.ts"]
