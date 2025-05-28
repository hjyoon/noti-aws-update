FROM docker.io/library/node:20

WORKDIR /app

COPY collect/ .
COPY .env .env

RUN npm i

CMD ["npm", "run", "start"]
