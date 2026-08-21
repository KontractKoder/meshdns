# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /meshdns ./cmd/meshdns

# Stage 2: Minimal runtime image (alpine works on Render; distroless pull is flaky)
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /meshdns /meshdns
COPY web/ /web/

EXPOSE 8080

ENTRYPOINT ["/meshdns"]
