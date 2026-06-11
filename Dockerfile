# ─────────────────────────────────────────────────────────
# Stage 1 — Build the Vue 3 PWA into backend/web/dist
# ─────────────────────────────────────────────────────────
FROM node:22-alpine AS web

# Match local dev CWD (`npm run dev` is run from frontend/) so vite's relative
# outDir `../backend/web/dist` lands at /app/backend/web/dist — a sibling of
# the go-build stage's WORKDIR /app, keeping cross-stage COPY paths symmetric.
WORKDIR /app/frontend

# Cache npm install layer
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci --no-audit --no-fund

# Build (output dir is configured in vite.config.ts → ../backend/web/dist)
COPY frontend/ ./
RUN npm run build

# ─────────────────────────────────────────────────────────
# Stage 2 — Build the Go server binary (embeds web/dist)
# ─────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS go-build

WORKDIR /app

# Cache go mod layer
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source + the freshly-built frontend dist
# (web stage writes it to /app/backend/web/dist).
COPY backend/ ./
COPY --from=web /app/backend/web/dist ./web/dist

# Build a static Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -buildid=" \
    -o /out/server \
    .

# ─────────────────────────────────────────────────────────
# Stage 3 — Minimal runtime image (distroless, no shell)
# ─────────────────────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

COPY --from=go-build /out/server /server

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/server"]
