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

CREATE TABLE IF NOT EXISTS resources (
    id SERIAL PRIMARY KEY,
    user_id integer NOT NULL,
    type integer NOT NULL,
    amount integer NOT NULL DEFAULT 0 CHECK (amount >= 0),

    UNIQUE(user_id, type),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
