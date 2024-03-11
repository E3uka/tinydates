-- for simplicity not dealing with foreign key constraints
CREATE TABLE IF NOT EXISTS "swipes" (
    "id" bigserial PRIMARY KEY,
    "swiper" integer NOT NULL,
    "swipee" integer NOT NULL,
    "decision" boolean NOT NULL
);
