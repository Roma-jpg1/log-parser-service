CREATE TABLE IF NOT EXISTS logs (
    id SERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    status TEXT NOT NULL,
    error_text TEXT NOT NULL DEFAULT '',
    nodes_count INT NOT NULL DEFAULT 0,
    ports_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nodes (
    id SERIAL PRIMARY KEY,
    log_id INT REFERENCES logs(id) ON DELETE CASCADE,
    source_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    UNIQUE (log_id, source_id)
);

CREATE TABLE IF NOT EXISTS ports (
    id SERIAL PRIMARY KEY,
    log_id INT REFERENCES logs(id) ON DELETE CASCADE,
    node_id INT REFERENCES nodes(id) ON DELETE CASCADE,
    source_id TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT
);

CREATE TABLE IF NOT EXISTS nodes_info (
    id SERIAL PRIMARY KEY,
    node_id INT REFERENCES nodes(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT
);

CREATE INDEX IF NOT EXISTS idx_nodes_log_id ON nodes(log_id);
CREATE INDEX IF NOT EXISTS idx_ports_log_id ON ports(log_id);
CREATE INDEX IF NOT EXISTS idx_ports_node_id ON ports(node_id);
CREATE INDEX IF NOT EXISTS idx_nodes_info_node_id ON nodes_info(node_id);
