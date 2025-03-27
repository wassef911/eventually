FROM golang:1.24-alpine

WORKDIR /app

RUN go install github.com/githubnemo/CompileDaemon@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ENTRYPOINT CompileDaemon --build="go build -o main cmd/server/main.go" --command=./main