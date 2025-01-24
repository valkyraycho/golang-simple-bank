FROM golang:1.23-alpine3.21 AS build

WORKDIR /app

COPY . .

RUN go build -o main main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=build /app/main .
COPY .env .
COPY db/migration ./db/migration

EXPOSE 8080

CMD [ "./main" ]