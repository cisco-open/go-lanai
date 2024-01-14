DROP TABLE IF EXISTS "public"."users";
CREATE TABLE friends
(
    id                  UUID NOT NULL DEFAULT gen_random_uuid(),
    first_name          STRING NOT NULL,
    last_name           STRING NOT NULL,
    CONSTRAINT          "primary" PRIMARY KEY (id ASC)
);