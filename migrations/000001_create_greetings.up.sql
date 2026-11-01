BEGIN;

CREATE TABLE greetings (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    language   TEXT        NOT NULL UNIQUE,
    message    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;
