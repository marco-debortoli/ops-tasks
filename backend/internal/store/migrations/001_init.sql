CREATE TABLE categories (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE CHECK (name <> ''),
    color      TEXT NOT NULL CHECK (color ~ '^#[0-9a-fA-F]{6}$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE projects (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL CHECK (name <> ''),
    category_id BIGINT REFERENCES categories (id) ON DELETE SET NULL,
    status      TEXT NOT NULL DEFAULT 'todo'
                CHECK (status IN ('todo', 'in_progress', 'waiting', 'complete')),
    due_date    DATE,
    notes       TEXT NOT NULL DEFAULT '',
    pinned      BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Pinning needs a category, and a completed project can't stay pinned.
    CHECK (NOT pinned OR (category_id IS NOT NULL AND status <> 'complete'))
);

-- One pinned project per category.
CREATE UNIQUE INDEX projects_one_pin_per_category ON projects (category_id) WHERE pinned;

CREATE TABLE tasks (
    id             BIGSERIAL PRIMARY KEY,
    name           TEXT NOT NULL CHECK (name <> ''),
    description    TEXT NOT NULL DEFAULT '',
    priority       SMALLINT CHECK (priority BETWEEN 1 AND 3),
    scheduled_date DATE,
    due_date       DATE,
    project_id     BIGINT REFERENCES projects (id) ON DELETE SET NULL,
    -- Only standalone tasks carry their own category; tasks in a project use the project's.
    category_id    BIGINT REFERENCES categories (id) ON DELETE SET NULL,
    -- completed_on is the (editable) date history and stats use; completed_at is the actual time.
    completed_on   DATE,
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (project_id IS NULL OR category_id IS NULL),
    CHECK ((completed_on IS NULL) = (completed_at IS NULL))
);

CREATE INDEX tasks_project_id ON tasks (project_id);
CREATE INDEX tasks_open ON tasks (priority, due_date) WHERE completed_on IS NULL;
CREATE INDEX tasks_completed_on ON tasks (completed_on) WHERE completed_on IS NOT NULL;

CREATE TABLE subtasks (
    id         BIGSERIAL PRIMARY KEY,
    task_id    BIGINT NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    name       TEXT NOT NULL CHECK (name <> ''),
    done       BOOLEAN NOT NULL DEFAULT false,
    position   INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX subtasks_task_id ON subtasks (task_id, position);
