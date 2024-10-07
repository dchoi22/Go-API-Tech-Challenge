FROM golang:1.23.1 AS builder

WORKDIR /app

ENV GOARCH=arm64
ENV GOOS=linux
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./cmd/api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8000

CMD ["./main"]