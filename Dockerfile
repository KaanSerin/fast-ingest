# ---- Build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/server       ./cmd/server
RUN go build -o /bin/migrate      ./cmd/migrate
RUN go build -o /bin/migrate-down ./cmd/migrate-down

# ---- Runtime stage ----
FROM alpine:3.23

WORKDIR /app

COPY --from=builder /bin/server        ./server
COPY --from=builder /bin/migrate       ./migrate
COPY --from=builder /bin/migrate-down  ./migrate-down
COPY migrations/                       ./migrations/

EXPOSE 8080

CMD ["./server"]
