ARG BASE_IMAGE=golang:1.24-alpine3.23
FROM ${BASE_IMAGE} AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG TARGETVARIANT=

WORKDIR /app

# Install build tools and libpcap using the container's package manager
RUN if command -v apk >/dev/null 2>&1; then \
			apk add --no-cache libpcap-dev build-base; \
		elif command -v apt-get >/dev/null 2>&1; then \
			apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y libpcap-dev build-essential pkg-config && rm -rf /var/lib/apt/lists/*; \
		else \
			echo "Unknown package manager; please ensure libpcap and build tools are present" >&2 && exit 1; \
		fi

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Ensure output dir
RUN mkdir -p /app

# Build with CGO enabled. TARGETVARIANT is used for GOARM when building for arm.
ENV CGO_ENABLED=1
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT} go build -trimpath -ldflags "-s -w" -o /app/main .