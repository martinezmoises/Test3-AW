CREATE TABLE reading_lists (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    created_by INTEGER REFERENCES users (id) ON DELETE CASCADE,
    books INTEGER[] DEFAULT '{}',
    status TEXT NOT NULL CHECK (status IN ('currently reading', 'completed')),
    created_at TIMESTAMP DEFAULT now()
);
