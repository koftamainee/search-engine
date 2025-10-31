CREATE DATABASE accounts_db;

CREATE USER accounts_user
WITH
  PASSWORD 'accounts_pass';

GRANT ALL PRIVILEGES ON DATABASE accounts_db TO accounts_user;

\connect accounts_db

CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT NOW (),
  last_login TIMESTAMP
);

CREATE TABLE IF NOT EXISTS search_history (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users (id) ON DELETE CASCADE,
  query TEXT NOT NULL,
  searched_at TIMESTAMP DEFAULT NOW ()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users (id) ON DELETE CASCADE,
  token TEXT NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT NOW (),
  expires_at TIMESTAMP
);

CREATE INDEX idx_search_history_user_id ON search_history (user_id);

ALTER DATABASE accounts_db OWNER TO accounts_user;
