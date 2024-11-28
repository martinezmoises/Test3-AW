CREATE TABLE tokens (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    hash BYTEA NOT NULL UNIQUE,
    scope TEXT NOT NULL,
    expiry TIMESTAMP NOT NULL
);

-- Index for efficient lookups by hash and scope
CREATE INDEX idx_tokens_hash_scope ON tokens (hash, scope);
