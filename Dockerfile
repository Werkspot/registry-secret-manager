FROM golang:1.20-alpine AS builder

RUN mkdir -p /app
WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./
RUN GOOS=linux go build -o registry-secret-manager

FROM alpine:latest
USER nobody

COPY --from=builder /app/registry-secret-manager /
COPY --from=builder /app/config.yml /

ENTRYPOINT ["/registry-secret-manager"]
