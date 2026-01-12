FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git tzdata
WORKDIR /app

COPY go.mod ./
COPY . .

# Clean and regenerate go.sum, then build
RUN go mod download && \
    go mod verify && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app

# Copy timezone data from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/app .
COPY --from=builder /app/permissions.yml .

# Set timezone to CET (GMT+1)
ENV TZ=CET

EXPOSE 8080
CMD ["./app"]
