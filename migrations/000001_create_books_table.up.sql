CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    authors TEXT[] NOT NULL,
    isbn VARCHAR(13) NOT NULL UNIQUE,
    publication_date DATE NOT NULL,
    genre TEXT NOT NULL,
    description TEXT NOT NULL,
    average_rating NUMERIC(3, 2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT now() NOT NULL,
    version INT NOT NULL DEFAULT 1
);
