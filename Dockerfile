# Build stage
FROM golang:1.24 AS builder

WORKDIR /src

COPY . .

# Install Bazelisk
RUN go install github.com/bazelbuild/bazelisk@latest

ENV PATH="/go/bin:${PATH}"

# Build the application with Bazel
RUN bazel build //cmd/server:server

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /src/bazel-bin/cmd/server/server ./server

EXPOSE 8080

USER 10001:10001

ENTRYPOINT ["./server"]
