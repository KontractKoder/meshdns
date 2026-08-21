# Single-stage build — avoids multi-stage pull issues on Render
FROM golang:1.24-alpine

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /meshdns ./cmd/meshdns

EXPOSE 8080

ENTRYPOINT ["/meshdns"]