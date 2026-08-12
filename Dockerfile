# Frontend build stage
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Backend build stage
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app

RUN apk add --no-cache git

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN go build -buildvcs=false -o server .

# Combined runtime image
FROM alpine:3.23

WORKDIR /app

RUN apk add --no-cache ca-certificates jq tzdata \
    && adduser -D -g '' appuser

COPY --from=backend-builder /app/server ./server
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

RUN chmod +x /usr/local/bin/docker-entrypoint.sh \
    && mkdir -p /app/logs \
    && chown -R appuser:appuser /app

USER appuser

EXPOSE 3002

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["./server", "-release=true"]
