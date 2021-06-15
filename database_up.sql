CREATE TABLE IF NOT EXISTS user (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(16) NOT NULL,
    email VARCHAR(255) NOT NULL,
    pass VARCHAR(255) NOT NULL,
    exp INT UNSIGNED NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,

    CONSTRAINT name_unique UNIQUE (name),
    CONSTRAINT email_unique UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS unit_type (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(32) NOT NULL,

    CONSTRAINT name_unique UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS unit_template (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    type_id INT UNSIGNED NOT NULL,
    name VARCHAR(32) NOT NULL,
    hp SMALLINT UNSIGNED NOT NULL CHECK (hp > 0),
    atk SMALLINT UNSIGNED NOT NULL CHECK (atk > 0),
    def SMALLINT UNSIGNED NOT NULL CHECK (def > 0),
    spd SMALLINT UNSIGNED NOT NULL CHECK (spd > 0),

    CONSTRAINT name_unique UNIQUE (name),
    FOREIGN KEY (type_id) REFERENCES unit_type(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS unit (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL,
    template_id INT UNSIGNED NOT NULL,
    level SMALLINT UNSIGNED NOT NULL DEFAULT 1 CHECK (level > 0 AND level <= 290),
    stars TINYINT UNSIGNED NOT NULL DEFAULT 1 CHECK (stars > 0 AND stars <= 10),
    is_locked TINYINT NOT NULL DEFAULT 0,

    FOREIGN KEY (user_id) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (template_id) REFERENCES unit_template(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS resource (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(32) NOT NULL,

    CONSTRAINT name_unique UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS user_resource (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL,
    resource_id INT UNSIGNED NOT NULL,
    amount INT UNSIGNED NOT NULL DEFAULT 0,

    CONSTRAINT user_id_resource_id_unique UNIQUE (user_id, resource_id),
    FOREIGN KEY (user_id) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (resource_id) REFERENCES resource(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS campaign (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL,
    level INT UNSIGNED NOT NULL DEFAULT 1,
    last_collected_at TIMESTAMP NOT NULL,

    CONSTRAINT user_id_unique UNIQUE (user_id),
    FOREIGN KEY (user_id) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS daily_quest (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    description VARCHAR(255) NOT NULL,
    required INT UNSIGNED NOT NULL
);

CREATE TABLE IF NOT EXISTS user_daily_quest (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL,
    daily_quest_id INT UNSIGNED NOT NULL,
    count INT UNSIGNED NOT NULL DEFAULT 0,

    CONSTRAINT user_id_daily_quest_id_unique UNIQUE (user_id, daily_quest_id),
    FOREIGN KEY (user_id) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (daily_quest_id) REFERENCES daily_quest(id) ON UPDATE CASCADE ON DELETE CASCADE
);

