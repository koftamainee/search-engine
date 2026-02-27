CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL UNIQUE,
    password    TEXT NOT NULL,
    is_admin    BOOLEAN DEFAULT FALSE,
    is_banned   BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMP DEFAULT NOW(),
    deleted_at  TIMESTAMP
);

CREATE TABLE search_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    query       TEXT NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE bookmarks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    url         TEXT NOT NULL,
    title       TEXT,
    description TEXT,
    created_at  TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, url)
);

CREATE INDEX idx_users_deleted_at          ON users(deleted_at);

CREATE INDEX idx_search_history_user_id    ON search_history(user_id);
CREATE INDEX idx_bookmarks_user_id         ON bookmarks(user_id);