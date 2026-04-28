CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short VARCHAR(50) NOT NULL,
    original VARCHAR(255) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_short ON urls(short);

CREATE INDEX IF NOT EXISTS idx_original ON urls(original);