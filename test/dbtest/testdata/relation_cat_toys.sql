INSERT INTO "test_cat_toys" ("cat_id", "toy_id")
VALUES ('612b71a3-ad84-4f11-b81b-9499925de8f4', 'e0474ad8-3ece-4c9c-a2f6-df0892fb4c12'),
       ('612b71a3-ad84-4f11-b81b-9499925de8f4', 'e4ae0e96-8f5c-4d75-8e39-17a8d2f582f4'),
       ('e109f3ca-3a60-43af-80dc-6637ea101e3d', 'e4ae0e96-8f5c-4d75-8e39-17a8d2f582f4'),
       ('e109f3ca-3a60-43af-80dc-6637ea101e3d', '92cc5c4c-5e7c-4724-b2ed-fdd316a5393b'),
       ('d2319447-3695-465a-882d-6d144182adb6', 'e0474ad8-3ece-4c9c-a2f6-df0892fb4c12'),
       ('d2319447-3695-465a-882d-6d144182adb6', '92cc5c4c-5e7c-4724-b2ed-fdd316a5393b')
ON CONFLICT DO NOTHING
RETURNING "cat_id","toy_id"