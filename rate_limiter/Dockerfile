# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /rate-limiter ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /rate-limiter /app/rate-limiter
USER nobody:nobody
EXPOSE 8080
ENTRYPOINT ["/app/rate-limiter"]
