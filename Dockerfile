# Development stage
FROM golang:1.26-alpine AS dev

RUN apk add --no-cache git tzdata

RUN go install github.com/bokwoon95/wgo@latest

WORKDIR /app

# Pre-download dependencies if go.mod/sum haven't changed
COPY go.mod go.sum ./
RUN go mod download

CMD ["wgo", "run", "cmd/firefly-importer/main.go"]

# Builder stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o firefly-importer cmd/firefly-importer/main.go

# Production stage
FROM alpine:latest AS prod

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app

COPY --from=builder /app/firefly-importer .

EXPOSE 8080

CMD ["/app/firefly-importer"]
