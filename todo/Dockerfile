FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o todo .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/todo ./todo
COPY --from=builder /app/web ./web

ENV TODO_PORT=7540

EXPOSE 7540

ENTRYPOINT ["./todo"]

