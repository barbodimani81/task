# Migrations image: the golang-migrate CLI plus this repo's SQL files.
# The CLI is built from source with only the Postgres driver and the file
# source, so the image carries none of the other drivers or their CVEs.
FROM golang:1.27 AS migrate
RUN CGO_ENABLED=0 go install -tags 'postgres file' -ldflags '-s -w -X main.Version=v4.20.1' \
    github.com/golang-migrate/migrate/v4/cmd/migrate@v4.20.1

# Same pinned distroless base as the app image.
FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329

COPY --from=migrate /go/bin/migrate /usr/local/bin/migrate
COPY *.sql /migrations/

USER 10001

ENTRYPOINT ["/usr/local/bin/migrate", "-path", "/migrations"]
CMD ["--help"]
