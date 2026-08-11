FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
# NEXT_PUBLIC_* vars are inlined into the client bundle at build time, not
# read from the container's runtime environment - so this has to be a
# build ARG, not a Deployment env var (see chart/livestream-web and
# .github/workflows/livestream-web-cicd.yml, which pass it here).
ARG NEXT_PUBLIC_LIVESTREAM_WS_URL
ENV NEXT_PUBLIC_LIVESTREAM_WS_URL=${NEXT_PUBLIC_LIVESTREAM_WS_URL}
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
# Next.js `output: "standalone"` (see next.config.js) traces and copies
# only the files needed to run `node server.js` - not the full
# node_modules tree.
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs
EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

CMD ["node", "server.js"]
