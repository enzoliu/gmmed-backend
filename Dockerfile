# base builder
FROM golang:1.23 AS builder
WORKDIR /app
RUN apt update && apt-get --no-install-recommends install -y libvips-dev pkg-config
COPY go.mod go.sum ./
COPY migrations/ ./migrations/
RUN go mod download
COPY . .

# base runtime
FROM debian:bookworm-slim AS runtime-cgo
RUN apt update && apt-get --no-install-recommends install -y libvips tzdata ca-certificates

# main
FROM builder AS main-builder
RUN CGO_ENABLED=1 GOOS=linux go build -o main "./cmd/server/."

FROM runtime-cgo AS main
COPY --from=main-builder /app/main /main
COPY --from=main-builder /app/migrations/ /migrations/
ENV TZ=Asia/Taipei
ENTRYPOINT ["/main"]
