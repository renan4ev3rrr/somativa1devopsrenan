# Build stage: keeps the final image small and free of source code.
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o lista-tarefas .

# Runtime stage: only the compiled application is shipped.
FROM alpine:3.20

RUN adduser -D -H appuser
WORKDIR /app
COPY --from=builder /app/lista-tarefas .
USER appuser

EXPOSE 8080
ENV PORT=8080
CMD ["./lista-tarefas"]
