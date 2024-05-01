-- Model C, full access attributes: tenant_id, tenant_path, owner
CREATE TABLE IF NOT EXISTS public.test_opa_model_c
(
    id            UUID        NOT NULL DEFAULT gen_random_uuid(),
    "tenant_name" TEXT      NOT NULL,
    "owner_name"  TEXT      NOT NULL,
    "value"       TEXT      NOT NULL,
    tenant_id     UUID        NULL,
    tenant_path   UUID[]      NULL,
    owner_id      UUID        NULL,
    created_at    TIMESTAMPTZ NULL,
    updated_at    TIMESTAMPTZ NULL,
    created_by    UUID        NULL,
    updated_by    UUID        NULL,
    PRIMARY KEY (id)
);