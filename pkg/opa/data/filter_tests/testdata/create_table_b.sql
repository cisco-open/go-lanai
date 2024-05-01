-- Model B, owner and sharing
CREATE TABLE IF NOT EXISTS public.test_opa_model_b
(
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    "owner_name" TEXT      NOT NULL,
    "value"      TEXT      NOT NULL,
    owner_id     UUID        NOT NULL,
    sharing      JSONB       NULL,
    created_at   TIMESTAMPTZ NULL,
    updated_at   TIMESTAMPTZ NULL,
    created_by   UUID        NULL,
    updated_by   UUID        NULL,
    deleted_at   TIMESTAMPTZ NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.test_opa_model_b_share
(
    id         UUID     NOT NULL DEFAULT gen_random_uuid(),
    res_id     UUID     NULL,
    user_id    UUID     NULL,
    username   TEXT   NOT NULL,
    operations TEXT[] NULL,
    PRIMARY KEY (id),
    CONSTRAINT "unique_res_user" UNIQUE (res_id, user_id)
);