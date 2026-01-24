# syntax=docker/dockerfile:1

# Build the manager binary
FROM golang:1.24 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /workspace

# Copy go mod files first for better layer caching
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy source
COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/

# Build with version info
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -a -ldflags="-w -s \
    -X main.version=${VERSION} \
    -X main.commit=${COMMIT} \
    -X main.date=${BUILD_DATE}" \
    -o manager cmd/main.go

# Runtime image
FROM gcr.io/distroless/static:nonroot

# OCI labels
LABEL org.opencontainers.image.title="Boilerr"
LABEL org.opencontainers.image.description="Kubernetes operator for managing Steam dedicated game servers"
LABEL org.opencontainers.image.source="https://github.com/CraightonH/boilerr"
LABEL org.opencontainers.image.url="https://github.com/CraightonH/boilerr"
LABEL org.opencontainers.image.documentation="https://github.com/CraightonH/boilerr/blob/main/README.md"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.vendor="CraightonH"

WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
