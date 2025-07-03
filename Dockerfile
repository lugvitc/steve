FROM golang:1.23-alpine AS builder
RUN apk add --no-cache build-base
WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 GOOS=`go env GOHOSTOS` GOARCH=`go env GOHOSTARCH` go build -o out/ubot -ldflags="-w -s" .

FROM alpine:latest AS app
RUN apk add tzdata --no-cache
WORKDIR /data
COPY --from=builder /app/out/ubot /app/ubot
CMD ["/app/ubot"]
