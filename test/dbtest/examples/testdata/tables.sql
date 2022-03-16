-- Table test_cats
CREATE TABLE IF NOT EXISTS security_clients
(
    id   uuid DEFAULT gen_random_uuid() NOT NULL,
    oauth_client_id text                           NOT NULL,
    CONSTRAINT "primary" PRIMARY KEY (id),
    CONSTRAINT idx_oauth_clint_id UNIQUE (oauth_client_id)
);

