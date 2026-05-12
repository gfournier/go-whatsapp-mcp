FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o discord-mcp ./discord-mcp

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && \
    addgroup -S mcp && \
    adduser -S mcp -G mcp

WORKDIR /app
COPY --from=builder /app/discord-mcp .

RUN chown mcp:mcp /app/discord-mcp
USER mcp

ENTRYPOINT ["./discord-mcp"]
