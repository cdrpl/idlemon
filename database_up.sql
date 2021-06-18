CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(16) NOT NULL,
    email VARCHAR(255) NOT NULL,
    pass VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    data JSONB NOT NULL,

    UNIQUE(name),
    UNIQUE(email)
);
