-- Model D-1
CREATE TABLE IF NOT EXISTS public.test_opa_model_d_user
(
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    "username" TEXT      NOT NULL,
    created_at TIMESTAMPTZ NULL,
    updated_at TIMESTAMPTZ NULL,
    created_by UUID        NULL,
    updated_by UUID        NULL,
    PRIMARY KEY (id)
);

-- Model D-2
CREATE TABLE IF NOT EXISTS public.test_opa_model_d
(
    id            UUID        NOT NULL DEFAULT gen_random_uuid(),
    "tenant_name" TEXT      NOT NULL,
    "value"       TEXT      NOT NULL,
    tenant_id     UUID        NULL,
    tenant_path   UUID[]      NULL,
    created_at    TIMESTAMPTZ NULL,
    updated_at    TIMESTAMPTZ NULL,
    created_by    UUID        NULL,
    updated_by    UUID        NULL,
    PRIMARY KEY (id)
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
    PRIMARY KEY (user_id, model_id)
);

-- Model D-1 and D-2 FK
ALTER TABLE public.test_opa_model_d_ref
    ADD CONSTRAINT fk_user_ref FOREIGN KEY (user_id) REFERENCES public.test_opa_model_d_user (id);
ALTER TABLE public.test_opa_model_d_ref
    ADD CONSTRAINT fk_model_ref FOREIGN KEY (model_id) REFERENCES public.test_opa_model_d (id);