-- idempotent table creation
CREATE TABLE IF NOT EXISTS "users" (
    "id" bigserial PRIMARY KEY,
    "email" varchar NOT NULL,
    "password" varchar NOT NULL,
    "name" varchar NOT NULL,
    "gender" varchar NOT NULL,
    "age" integer NOT NULL,
    UNIQUE(id, email)
);
