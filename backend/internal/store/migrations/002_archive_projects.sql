-- An archived project is hidden from the dashboard, along with its open tasks.
ALTER TABLE projects ADD COLUMN archived_at TIMESTAMPTZ;
ALTER TABLE projects ADD CONSTRAINT projects_archived_not_pinned CHECK (NOT pinned OR archived_at IS NULL);
