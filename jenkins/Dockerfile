# Jenkins with the toolchain the pipeline needs: Go (gofmt, and the host SDK
# that MODULE.bazel's go_sdk.host() expects), a C compiler for rules_go,
# and Bazelisk exposed as `bazel` (it reads .bazelversion).
FROM golang:1.24 AS tools
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

USER jenkins
RUN jenkins-plugin-cli --plugins gerrit-trigger git workflow-aggregator timestamper ssh-credentials
