# Build stage
FROM golang:1.24 AS builder

WORKDIR /src

COPY . .

# Install Bazelisk
RUN go install github.com/bazelbuild/bazelisk@v1.29.0

ENV PATH="/go/bin:${PATH}"

# Build the application with Bazel
RUN bazelisk build //cmd/server:server && \
    cp "$(bazelisk cquery --output=files //cmd/server:server)" /server

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /server ./server

EXPOSE 8080

USER 10001:10001

ENTRYPOINT ["./server"]
