FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web ./
RUN npm run build

FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /keeper ./cmd/keeper

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /keeper /keeper
COPY --from=web /web/dist /static
ENV APP_STATIC_DIR=/static
ENTRYPOINT ["/keeper"]
