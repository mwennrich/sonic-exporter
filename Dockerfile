# ===========
# Build stage
# ===========
FROM golang:1.23-alpine  AS builder

WORKDIR /code

ENV CGO_ENABLED=0

# Pre-install dependencies to cache them as a separate image layer
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . /code
RUN go build -o sonic-exporter ./cmd/sonic-exporter/main.go

# ===========
# Final stage
# ===========
FROM scratch

WORKDIR /app

COPY --from=builder /code/sonic-exporter /

ENTRYPOINT [ "/sonic-exporter" ]
