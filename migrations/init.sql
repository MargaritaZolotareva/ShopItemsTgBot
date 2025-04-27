CREATE TABLE balloon
(
    id    SERIAL PRIMARY KEY,
    sku   INT NOT NULL,
    name  VARCHAR(128) NOT NULL,
    balloon_price INT NOT NULL,
    helium_portions FLOAT NOT NULL DEFAULT 1,
    hi_float INT NOT NULL DEFAULT 0
);