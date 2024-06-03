CREATE TABLE IF NOT EXISTS friends
(
    id                  UUID NOT NULL DEFAULT gen_random_uuid(),
    first_name          TEXT NOT NULL,
    last_name           TEXT NOT NULL,
    created_at          timestamp with time zone,
    updated_at          timestamp with time zone,
    created_by          TEXT,
    updated_by          TEXT,
    CONSTRAINT          "primary" PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_friends_name ON friends (first_name, last_name);