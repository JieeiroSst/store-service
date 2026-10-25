FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
ARG VITE_APK_URL
ENV VITE_APK_URL=$VITE_APK_URL
ARG VITE_ADSENSE_CLIENT
ENV VITE_ADSENSE_CLIENT=$VITE_ADSENSE_CLIENT
ARG VITE_ADSENSE_SLOT
ENV VITE_ADSENSE_SLOT=$VITE_ADSENSE_SLOT
RUN npm run build

FROM nginxinc/nginx-unprivileged:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 3000
USER 101
