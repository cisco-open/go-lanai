CREATE TABLE IF NOT EXISTS data_encryption_test
(
    id    int   NOT NULL,
    name  text  NOT NULL,
    value jsonb NOT NULL,
    CONSTRAINT "primary" PRIMARY KEY (id),
    CONSTRAINT idx_record_name UNIQUE (name)
);

INSERT INTO "data_encryption_test" ("id", "name", "value")
VALUES (675153534251466753, 'v1_plain_map', '{"alg": "p", "d": {"key1": "value1", "key2": 2}, "kid": "d034a284-172f-46c3-aead-e7cfb2f78ddc", "v": 1}'),
       (675153534266441729, 'v2_plain_map', '{"alg": "p", "d": {"key1": "value1", "key2": 2}, "kid": "d034a284-172f-46c3-aead-e7cfb2f78ddc", "v": 2}'),
       (675154428705505281, 'v1_mock_map', '{"alg": "e", "d": "d034a284-172f-46c3-aead-e7cfb2f78ddc:{\"key1\":\"value1\",\"key2\":2}", "kid": "d034a284-172f-46c3-aead-e7cfb2f78ddc", "v": 1}'),
       (675154428730507265, 'v2_mock_map', '{"alg": "e", "d": "d034a284-172f-46c3-aead-e7cfb2f78ddc:{\"key1\":\"value1\",\"key2\":2}", "kid": "d034a284-172f-46c3-aead-e7cfb2f78ddc", "v": 2}'),
       (675171350698229761, 'v2_invalid_plain_map', '{"alg": "p", "d": "invalid", "kid": "d034a284-172f-46c3-aead-e7cfb2f78ddc", "v": 2}'),
       (675171350734307329, 'v2_invalid_mock_map', '{"alg": "e", "d": "invalid", "kid": "d034a284-172f-46c3-aead-e7cfb2f78ddc", "v": 2}')
ON CONFLICT DO NOTHING
RETURNING "id"