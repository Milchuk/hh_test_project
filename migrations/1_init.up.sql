CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE goods (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    name TEXT NOT NULL,
    description TEXT,
    priority INTEGER NOT NULL,
    removed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_id ON goods(project_id);
CREATE INDEX idx_goods_id ON goods(id);
CREATE INDEX idx_goods_project_id ON goods(project_id);
CREATE INDEX idx_goods_name ON goods(name);

CREATE OR REPLACE FUNCTION set_priority() RETURNS TRIGGER AS $$
BEGIN
    SELECT COALESCE(MAX(priority), 0) + 1 INTO NEW.priority 
    FROM goods 
    WHERE project_id = NEW.project_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_goods_priority
BEFORE INSERT ON goods
FOR EACH ROW EXECUTE FUNCTION set_priority();

INSERT INTO projects (name) VALUES ('Первая запись');