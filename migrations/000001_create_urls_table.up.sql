CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short VARCHAR(50) NOT NULL,
    original VARCHAR(255) NOT NULL
);

CREATE INDEX idx_short ON urls(short);

CREATE INDEX idx_original ON urls(original);