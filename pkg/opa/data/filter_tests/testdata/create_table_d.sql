-- Model D-1
CREATE TABLE IF NOT EXISTS public.test_opa_model_d_user
(
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    "username" STRING      NOT NULL,
    created_at TIMESTAMPTZ NULL,
    updated_at TIMESTAMPTZ NULL,
    created_by UUID        NULL,
    updated_by UUID        NULL,
    CONSTRAINT "primary" PRIMARY KEY (id ASC),
    INDEX      idx_username(username ASC),
    FAMILY     "primary"(id, username, created_at, updated_at, created_by, updated_by)
);

-- Model D-2
CREATE TABLE IF NOT EXISTS public.test_opa_model_d
(
    id            UUID        NOT NULL DEFAULT gen_random_uuid(),
    "tenant_name" STRING      NOT NULL,
    "value"       STRING      NOT NULL,
    tenant_id     UUID        NULL,
    tenant_path   UUID[]      NULL,
    created_at    TIMESTAMPTZ NULL,
    updated_at    TIMESTAMPTZ NULL,
    created_by    UUID        NULL,
    updated_by    UUID        NULL,
    CONSTRAINT "primary" PRIMARY KEY (id ASC),
    INVERTED      INDEX idx_tenant_path (tenant_path),
    INDEX         idx_tenant_id(tenant_id ASC),
    FAMILY        "primary"(id, tenant_name, value, tenant_id, tenant_path, created_at, updated_at, created_by, updated_by)
);

-- Model D-1 and D-2 Ref Table
CREATE TABLE IF NOT EXISTS public.test_opa_model_d_ref
(
    user_id    UUID        NOT NULL DEFAULT gen_random_uuid(),
    model_id   UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NULL,
    updated_at TIMESTAMPTZ NULL,
    created_by UUID        NULL,
    updated_by UUID        NULL,
    CONSTRAINT "primary" PRIMARY KEY (user_id ASC, model_id ASC),
    INDEX      idx_user_id(user_id ASC),
    INDEX      idx_model_id(model_id ASC),
    FAMILY     "primary"(user_id, model_id, created_at, updated_at, created_by, updated_by)
);

-- Model D-1 and D-2 FK
ALTER TABLE public.test_opa_model_d_ref
    ADD CONSTRAINT fk_user_ref FOREIGN KEY (user_id) REFERENCES public.test_opa_model_d_user (id);
ALTER TABLE public.test_opa_model_d_ref
    ADD CONSTRAINT fk_model_ref FOREIGN KEY (model_id) REFERENCES public.test_opa_model_d (id);