FROM golang:1.24.1 AS builder
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/api ./server/main.go

# Stage 2: Minimal image
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /app/api /app/api
EXPOSE 8080
CMD ["/app/api"]