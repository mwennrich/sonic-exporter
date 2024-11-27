# ===========
# Build stage
# ===========
FROM golang:1.23-alpine  AS builder

WORKDIR /code

# Pre-install dependencies to cache them as a separate image layer
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . /code
RUN go build -o sonic-exporter ./cmd/sonic-exporter/main.go

# ===========
# Final stage
# ===========
FROM alpine:3.20

WORKDIR /app
RUN apk --no-cache add curl

COPY --from=builder /code/sonic-exporter .

CMD [ "./sonic-exporter" ]
