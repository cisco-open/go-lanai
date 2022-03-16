-- Table test_cats
CREATE TABLE IF NOT EXISTS test_cats
(
    id   uuid DEFAULT gen_random_uuid() NOT NULL,
    name text                           NOT NULL,
    CONSTRAINT "primary" PRIMARY KEY (id),
    CONSTRAINT idx_cats_name UNIQUE (name)
);

-- Table test_toys
CREATE TABLE IF NOT EXISTS test_toys
(
    id   uuid DEFAULT gen_random_uuid() NOT NULL,
    what text                           NOT NULL,
    CONSTRAINT "primary" PRIMARY KEY (id)
);

-- Table test_cat_toys for many-to-many between api_keys and api_scopes
CREATE TABLE IF NOT EXISTS test_cat_toys
(
    cat_id uuid NOT NULL,
    toy_id uuid NOT NULL,
    CONSTRAINT "primary" PRIMARY KEY (cat_id, toy_id),
    CONSTRAINT fk_test_cat_toys_cat FOREIGN KEY (cat_id) REFERENCES test_cats,
    CONSTRAINT fk_test_cat_toys_toy FOREIGN KEY (toy_id) REFERENCES test_toys
);
