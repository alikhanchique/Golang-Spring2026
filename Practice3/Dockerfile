FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o myapp ./cmd/api/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /app/myapp .
COPY --from=builder /app/database/migrations ./database/migrations

ENTRYPOINT ["./myapp"]