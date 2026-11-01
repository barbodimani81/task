# Build stage
FROM golang:1.27 AS builder

WORKDIR /src

# Install Bazelisk
RUN go install github.com/bazelbuild/bazelisk@v1.29.0

ENV PATH="/go/bin:${PATH}"

COPY . .

# The cache mount keeps Bazel's downloads between builds, so CI doesn't start from zero.
RUN --mount=type=cache,target=/root/.cache \
    bazelisk build //cmd/server:server && \
    cp "$(bazelisk cquery --output=files //cmd/server:server)" /server

# Runtime stage
FROM debian:bookworm-slim

# The commit this image is built from; CI passes it with --build-arg.
ARG VERSION=dev
ENV APP_VERSION=${VERSION}

WORKDIR /app

COPY --from=builder /server ./server

EXPOSE 8080

USER 10001:10001

ENTRYPOINT ["./server"]
