FROM golang:1.23-alpine3.21 AS build

RUN apk add curl

WORKDIR /app

COPY . .
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz

RUN go build -o main main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=build /app/main .
COPY --from=build /app/migrate ./migrate
COPY .env .
COPY db/migration ./migration
COPY start.sh .

EXPOSE 8080

ENTRYPOINT [ "./start.sh" ]

CMD [ "./main" ]