
CREATE TABLE IF NOT EXISTS "product" (
    "id" CHAR(36) NOT NULL,
    "title" VARCHAR(100) NOT NULL,
    "created_by" VARCHAR(100) NOT NULL,
    "updated_by" VARCHAR(100) NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL,
    "updated_at" TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS "variant" (
    "id" CHAR(36) NOT NULL,
    "product_id" VARCHAR(100) NOT NULL,
    "title" VARCHAR(100) NOT NULL,
    "image" VARCHAR(100) NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS "image" (
    "id" CHAR(36) NOT NULL,
    "product_id" VARCHAR(100) NOT NULL,
    "is_main" BOOL NOT NULL,
    "src" VARCHAR(100) NOT NULL,
    "alt" VARCHAR(100) NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id)
);
