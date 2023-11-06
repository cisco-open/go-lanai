-- Model C, full access attributes: tenant_id, tenant_path, owner
CREATE TABLE IF NOT EXISTS public.test_opa_model_c
(
    id            UUID        NOT NULL DEFAULT gen_random_uuid(),
    "tenant_name" STRING      NOT NULL,
    "owner_name"  STRING      NOT NULL,
    "value"       STRING      NOT NULL,
    tenant_id     UUID        NULL,
    tenant_path   UUID[]      NULL,
    owner_id      UUID        NULL,
    created_at    TIMESTAMPTZ NULL,
    updated_at    TIMESTAMPTZ NULL,
    created_by    UUID        NULL,
    updated_by    UUID        NULL,
    CONSTRAINT "primary" PRIMARY KEY (id ASC),
    INVERTED      INDEX idx_tenant_path (tenant_path),
    INDEX         idx_tenant_id(tenant_id ASC),
    INDEX         idx_owner_id(owner_id ASC),
    FAMILY        "primary"(id, tenant_name, value, tenant_id, tenant_path, created_at, updated_at, created_by, updated_by)
);