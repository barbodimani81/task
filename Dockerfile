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

# Runtime stage: distroless has no shell and no package manager, only what a
# static binary needs (CA certificates, time zones). Pinned by digest.
FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329

# The commit this image is built from; CI passes it with --build-arg.
ARG VERSION=dev
ENV APP_VERSION=${VERSION}

WORKDIR /app

COPY --from=builder /server ./server

EXPOSE 8080

USER 10001:10001

ENTRYPOINT ["./server"]
