# Jenkins with the toolchain the pipeline needs: Go (gofmt, and the host SDK
# that MODULE.bazel's go_sdk.host() expects), a C compiler for rules_go,
# and Bazelisk exposed as `bazel` (it reads .bazelversion).
FROM golang:1.27 AS tools
RUN GOBIN=/out go install github.com/bazelbuild/bazelisk@v1.29.0

FROM jenkins/jenkins:lts-jdk21

USER root
RUN apt-get update \
 && apt-get install -y --no-install-recommends gcc libc6-dev \
 && rm -rf /var/lib/apt/lists/*
COPY --from=tools /usr/local/go /usr/local/go
COPY --from=tools /out/bazelisk /usr/local/bin/bazelisk
RUN ln -s /usr/local/bin/bazelisk /usr/local/bin/bazel
ENV PATH="/usr/local/go/bin:${PATH}"

# Docker CLI and buildx to build and push images. The daemon is the separate
# docker:dind service in docker-compose.yaml.
COPY --from=docker:29-cli /usr/local/bin/docker /usr/local/bin/docker
COPY --from=docker:29-cli /usr/local/libexec/docker/cli-plugins/docker-buildx /usr/local/libexec/docker/cli-plugins/docker-buildx

# kubectl, only for `kubectl kustomize`; CI never talks to a cluster.
ARG KUBECTL_VERSION=v1.34.1
ARG TARGETARCH
RUN curl -fsSLo /usr/local/bin/kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl" \
 && echo "$(curl -fsSL "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl.sha256")  /usr/local/bin/kubectl" | sha256sum -c - \
 && chmod +x /usr/local/bin/kubectl

# Trivy finds known vulnerabilities in images; cosign signs them. Both are single static
# binaries, copied out of their makers' pinned images.
COPY --from=aquasec/trivy:0.76.1 /usr/local/bin/trivy /usr/local/bin/trivy
COPY --from=ghcr.io/sigstore/cosign/cosign:v2.6.5 /ko-app/cosign /usr/local/bin/cosign

USER jenkins
RUN jenkins-plugin-cli --plugins gerrit-trigger git workflow-aggregator timestamper ssh-credentials
