ARG BASE_IMAGE=golang:1.24-alpine3.23
FROM ${BASE_IMAGE} AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG TARGETVARIANT=

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Ensure output dir
RUN mkdir -p /app

# Static build (no CGO). TARGETVARIANT is used for GOARM when building for arm.
ENV CGO_ENABLED=0
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT} \
	go build \
		-tags "netgo,osusergo" \
		-ldflags "-s -w -extldflags '-static'" \
		-trimpath \
		-o /app/main \
		.