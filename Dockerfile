FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]
