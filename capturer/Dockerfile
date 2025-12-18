FROM golang:1.26rc1-alpine3.23 AS builder

WORKDIR /app

RUN apk add --no-cache libpcap-dev build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /app/main .