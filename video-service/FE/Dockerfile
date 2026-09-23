FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# Leave empty to call the API on the same origin (the ingress routes /api to
# the backend). Set it only if the API lives on another host; it is inlined
# into the bundle at build time.
ARG REACT_APP_API_URL=""
ENV REACT_APP_API_URL=${REACT_APP_API_URL}
RUN npm run build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/build /usr/share/nginx/html
EXPOSE 80
