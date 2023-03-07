FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /app ./...

FROM gcr.io/distroless/static:latest
WORKDIR /app
COPY --from=builder /app/commit /app/commit

EXPOSE 8080
CMD ["/app/commit"]
