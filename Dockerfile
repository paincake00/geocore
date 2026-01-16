FROM golang:1.25.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o geocore cmd/geocore/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/geocore .
# Copy migrations if needed for runtime, though usually we might run them separately
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./geocore"]
