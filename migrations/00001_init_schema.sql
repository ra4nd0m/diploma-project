-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE "user" (
    id UUID PRIMARY KEY,
    display_name TEXT NOT NULL,
    preferences JSONB NOT NULL DEFAULT '{}' :: jsonb
);

CREATE TABLE cohort (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    owner UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_cohort_owner FOREIGN KEY (owner) REFERENCES "user"(id) ON DELETE RESTRICT
);

CREATE TABLE user_cohort (
    user_id UUID NOT NULL,
    cohort_id UUID NOT NULL,
    joined_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_user_cohort PRIMARY KEY (user_id, cohort_id),
    CONSTRAINT fk_user_cohort_user FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_cohort_cohort FOREIGN KEY (cohort_id) REFERENCES cohort(id) ON DELETE CASCADE
);

CREATE INDEX idx_cohort_owner ON cohort(owner);

CREATE INDEX idx_user_cohort_cohort_id ON user_cohort(cohort_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_cohort_cohort_id;

DROP INDEX IF EXISTS idx_cohort_owner;

DROP TABLE IF EXISTS user_cohort;

DROP TABLE IF EXISTS cohort;

DROP TABLE IF EXISTS "user";