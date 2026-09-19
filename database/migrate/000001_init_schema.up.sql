-- Active: 1746620303437@@8.153.101.141@3306@url_db
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS urls (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    original_url VARCHAR(2048) NOT NULL,
    short_code VARCHAR(100) NOT NULL UNIQUE,
    is_custom BOOLEAN NOT NULL DEFAULT FALSE,
    views INT NOT NULL DEFAULT 0,
    expired_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_expired_at_short_code ON urls(short_code, expired_at);
CREATE INDEX idx_expired_at_user_id ON urls(user_id, expired_at);

-- CREATE TABLE IF NOT EXISTS users (
--     id BIGINT AUTO_INCREMENT PRIMARY KEY,
--     email VARCHAR(255) NOT NULL UNIQUE,
--     password_hash TEXT NOT NULL,
--     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
-- );

-- CREATE INDEX idx_email ON users(email);

-- CREATE TABLE IF NOT EXISTS urls (
--     id BIGINT AUTO_INCREMENT PRIMARY KEY,
--     user_id BIGINT NOT NULL,
--     original_url TEXT NOT NULL,
--     short_code VARCHAR(100) NOT NULL UNIQUE,
--     is_custom BOOLEAN NOT NULL DEFAULT FALSE,
--     views INT NOT NULL DEFAULT 0,
--     expired_at TIMESTAMP NOT NULL,
--     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
-- );

-- CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
-- CREATE INDEX IF NOT EXISTS idx_expired_at_short_code ON urls(short_code, expired_at);
-- CREATE INDEX IF NOT EXISTS idx_user_id ON urls(user_id);
-- CREATE INDEX IF NOT EXISTS idx_expired_at_user_id ON urls(user_id, expired_at);
