FROM node:24-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.27-alpine AS api
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
LABEL org.opencontainers.image.source=https://github.com/marco-debortoli/ops-tasks
COPY --from=api /out/server /app/server
COPY --from=web /web/build /app/web
ENV STATIC_DIR=/app/web ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/app/server"]
