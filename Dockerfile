# --- frontend ---
FROM node:22-alpine AS fe
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npx nuxt generate

# --- backend (embeds the frontend) ---
FROM golang:1.26-alpine AS be
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN rm -rf internal/web/dist
COPY --from=fe /app/frontend/.output/public ./internal/web/dist
RUN CGO_ENABLED=0 go build -trimpath -o /out/server ./cmd/server \
 && CGO_ENABLED=0 go build -trimpath -o /out/migrate ./cmd/migrate

# --- runtime ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=be /out/ /app/
COPY backend/migrations ./migrations
ENV GIN_MODE=release
CMD ["sh", "-c", "./migrate up && ./server"]