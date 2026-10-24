FROM node:22

WORKDIR /app

COPY . .

RUN npm install

EXPOSE 1234

USER 1000
CMD ["npm", "start"]