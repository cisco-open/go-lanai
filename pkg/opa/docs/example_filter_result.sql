UPDATE "test_opa_model_a"
SET "owner_id"='a5aaa07a-e7d7-4f66-bec8-1e651badacbd',
    "updated_at"='2023-06-26 12:46:58.236'
WHERE ("test_opa_model_a"."tenant_id" = 'd8423acc-28cb-4209-95d6-089de7fb27ef' OR
       "test_opa_model_a"."tenant_path" @> '{"d8423acc-28cb-4209-95d6-089de7fb27ef"}' OR
       ("test_opa_model_a"."owner_id" = '595959e4-8803-4ab1-8acf-acfb92bb7322' AND
        "test_opa_model_a"."tenant_id" = 'd8423acc-28cb-4209-95d6-089de7fb27ef') OR
       ("test_opa_model_a"."owner_id" = '595959e4-8803-4ab1-8acf-acfb92bb7322' AND
        "test_opa_model_a"."tenant_path" @> '{"d8423acc-28cb-4209-95d6-089de7fb27ef"}'))
  AND "test_opa_model_a"."deleted_at" IS NULL
  AND "id" = '567ff75a-354f-4a3a-affe-4ced6f876864';

UPDATE "test_opa_model_a"
SET "value"='Updated',
    "updated_at"='2023-06-26 12:46:58.23'
WHERE (("test_opa_model_a"."owner_id" = '595959e4-8803-4ab1-8acf-acfb92bb7322' AND
        "test_opa_model_a"."tenant_id" = 'd8423acc-28cb-4209-95d6-089de7fb27ef') OR
       ("test_opa_model_a"."owner_id" = '595959e4-8803-4ab1-8acf-acfb92bb7322' AND
        "test_opa_model_a"."tenant_path" @> '{"d8423acc-28cb-4209-95d6-089de7fb27ef"}'))
  AND "test_opa_model_a"."deleted_at" IS NULL
  AND "id" = '957785c6-d75a-47cd-a3dd-e29f1219afd7';

SELECT *
FROM "my_resource"
WHERE ("my_resource"."owner_id" = '595959e4-8803-4ab1-8acf-acfb92bb7322' AND
       "my_resource"."tenant_id" = 'd8423acc-28cb-4209-95d6-089de7fb27ef')
   OR ("my_resource"."owner_id" = '595959e4-8803-4ab1-8acf-acfb92bb7322' AND
       "my_resource"."tenant_path" @> '{"d8423acc-28cb-4209-95d6-089de7fb27ef"}')
   OR ("sharing" -> '595959e4-8803-4ab1-8acf-acfb92bb7322' @> '"read"' AND
       "my_resource"."tenant_id" = 'd8423acc-28cb-4209-95d6-089de7fb27ef')
   OR ("sharing" -> '595959e4-8803-4ab1-8acf-acfb92bb7322' @> '"read"' AND
       "my_resource"."tenant_path" @> '{"d8423acc-28cb-4209-95d6-089de7fb27ef"}');