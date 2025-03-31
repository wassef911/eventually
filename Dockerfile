FROM golang:1.24-alpine as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o main  cmd/server/main.go

FROM alpine:latest as run

WORKDIR /app
COPY --from=build /app/main /app
CMD ["./main"]
