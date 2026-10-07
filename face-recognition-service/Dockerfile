FROM node:20-alpine

WORKDIR /app

ENV NODE_ENV=production

COPY package.json package-lock.json ./
RUN npm ci --omit=dev

COPY src ./src
COPY public ./public

USER node

EXPOSE 1234

CMD ["node", "src/server.js"]
