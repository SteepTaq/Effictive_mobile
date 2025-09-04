FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/service

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY --from=builder /app/.env.example /app/.env
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/static /app/static
COPY --from=builder /app/static/swagger.json /app/swagger.json
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/server"]


