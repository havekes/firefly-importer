FROM golang:1.26-alpine

RUN apk add --no-cache git tzdata

RUN go install github.com/bokwoon95/wgo@latest

WORKDIR /app

# Pre-download dependencies if go.mod/sum haven't changed
COPY go.mod go.sum ./
RUN go mod download

CMD ["wgo", "run", "cmd/firefly-importer/main.go"]
