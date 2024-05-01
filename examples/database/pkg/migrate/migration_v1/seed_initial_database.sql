DROP TABLE IF EXISTS "public"."users";
CREATE TABLE friends
(
    id                  UUID NOT NULL DEFAULT gen_random_uuid(),
    first_name          TEXT NOT NULL,
    last_name           TEXT NOT NULL,
    CONSTRAINT          "primary" PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_friends_name ON friends (first_name, last_name);