# Stage 1: Build the Go binary
FROM golang:1.23.6 AS builder


WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the agent binary
RUN CGO_ENABLED=0 GOOS=linux go build -o agent ./cmd/agent/main.go

# Stage 2: Minimal runtime container
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/agent /agent
COPY bpf/ /bpf/


ENTRYPOINT ["/agent"]
