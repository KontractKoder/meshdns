# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /meshdns ./cmd/meshdns

# Stage 2: Minimal runtime image
FROM gcr.io/distroless/static-debian12

COPY --from=builder /meshdns /meshdns
COPY web/ /web/

EXPOSE 8080

ENTRYPOINT ["/meshdns"]
