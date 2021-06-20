CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(16) NOT NULL,
    email VARCHAR(255) NOT NULL,
    pass VARCHAR(255) NOT NULL,
    exp integer NOT NULL DEFAULT 0 CHECK (exp >= 0),
    created_at TIMESTAMP NOT NULL,

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

CREATE TABLE IF NOT EXISTS campaign (
    id SERIAL PRIMARY KEY,
    user_id integer NOT NULL,
    level integer NOT NULL DEFAULT 1 CHECK (level >= 1),
    last_collected_at TIMESTAMP NOT NULL,

    UNIQUE(user_id),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS daily_quest_progress (
    id SERIAL PRIMARY KEY,
    user_id integer NOT NULL,
    daily_quest_id integer NOT NULL,
    count integer NOT NULL DEFAULT 0,
    last_completed_at TIMESTAMP NOT NULL,

    UNIQUE(user_id, daily_quest_id),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS units (
    id SERIAL PRIMARY KEY,
    user_id integer NOT NULL,
    template integer NOT NULL,
    level integer NOT NULL DEFAULT 1 CHECK (level >= 1),
    stars integer NOT NULL DEFAULT 1 CHECK (stars >= 1 AND stars <= 10),
    is_locked boolean NOT NULL DEFAULT FALSE,

    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

