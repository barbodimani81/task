# Migrations image: the golang-migrate CLI plus this repo's SQL files.
FROM migrate/migrate:v4.18.3

COPY *.sql /migrations/

USER 10001

ENTRYPOINT ["migrate", "-path", "/migrations"]
CMD ["--help"]
