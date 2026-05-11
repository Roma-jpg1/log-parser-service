CREATE TABLE IF NOT EXISTS logs (
    id SERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS nodes (
    id SERIAL PRIMARY KEY,
    log_id INT REFERENCES logs(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL
    );

CREATE TABLE IF NOT EXISTS ports (
    id SERIAL PRIMARY KEY,
    node_id INT REFERENCES nodes(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    status TEXT
    );

CREATE TABLE IF NOT EXISTS nodes_info (
    id SERIAL PRIMARY KEY,
    node_id INT REFERENCES nodes(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT
    );