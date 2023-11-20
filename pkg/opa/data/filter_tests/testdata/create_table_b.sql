-- Model B, owner and sharing
CREATE TABLE IF NOT EXISTS public.test_opa_model_b
(
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    "owner_name" STRING      NOT NULL,
    "value"      STRING      NOT NULL,
    owner_id     UUID        NOT NULL,
    sharing      JSONB       NULL,
    created_at   TIMESTAMPTZ NULL,
    updated_at   TIMESTAMPTZ NULL,
    created_by   UUID        NULL,
    updated_by   UUID        NULL,
    deleted_at   TIMESTAMPTZ NULL,
    CONSTRAINT "primary" PRIMARY KEY (id ASC),
    INDEX        idx_owner_id(owner_id ASC) STORING (sharing),
    INVERTED     INDEX idx_sharing (sharing),
    FAMILY       "primary"(id, value, created_at, updated_at, created_by, updated_by, deleted_at)
);

CREATE TABLE IF NOT EXISTS public.test_opa_model_b_share
(
    id         UUID     NOT NULL DEFAULT gen_random_uuid(),
    res_id     UUID     NULL,
    user_id    UUID     NULL,
    username   STRING   NOT NULL,
    operations STRING[] NULL,
    CONSTRAINT "primary" PRIMARY KEY (id ASC),
    CONSTRAINT "unique_res_user" UNIQUE (res_id, user_id),
    INVERTED   INDEX idx_operations (operations),
    INDEX      idx_res_id(res_id ASC),
    INDEX      idx_user_id(user_id ASC),
    FAMILY     "primary"(id, res_id, user_id, username, operations)
);