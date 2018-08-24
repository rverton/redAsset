CREATE TABLE hosts (
    id SERIAL PRIMARY KEY,
    hostname TEXT,
    apps JSONB
)
