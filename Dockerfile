# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS build
WORKDIR /app

# Install git (needed for module downloads)
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -o /app/bin/server ./cmd/server

# --- Runtime image ---
FROM alpine:3.21
WORKDIR /app

COPY --from=build /app/bin/server /app/bin/server

EXPOSE 8080

ENTRYPOINT ["/app/bin/server"]
