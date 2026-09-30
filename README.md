# ops · tasks

A single-user task manager with a terminal / ops look. SvelteKit (static SPA, Tailwind) on
the front, Go + Postgres behind it. Design decisions live in
[DESIGN_DECISIONS.md](DESIGN_DECISIONS.md).

## Run it

```bash
make up          # builds the image and starts Postgres + app on http://localhost:8080
```

Set `POSTGRES_PASSWORD` (and optionally `APP_PORT`) in a `.env` file for a real deployment.
`docker compose down -v` wipes the database.

## Develop

```bash
make db          # Postgres on localhost:5432
make api         # Go API on :8080 (uses local Go if installed, otherwise Docker)
make web         # Vite dev server on :5173, proxies /api to :8080
make check       # go vet, go test, svelte-check
```

## Layout

- `backend/cmd/server`: entry point. Env: `DATABASE_URL` (required), `ADDR` (`:8080`),
  `STATIC_DIR` (built frontend), `APP_TIMEZONE` (`America/Vancouver`, where "today" is decided).
- `backend/internal/store`: SQL, migrations (embedded, applied on start), and the domain
  rules: pinning, category inheritance, Today grouping, stats.
- `backend/internal/api`: JSON API under `/api/`, plus serving the SPA with an `index.html` fallback.
- `web/src/lib/components`: dashboard panels and the task / project / category / history modals.
  Theme tokens are in `web/src/app.css`.

## API

| Method | Path | |
| --- | --- | --- |
| GET | `/api/dashboard` | everything the dashboard shows, grouped server-side |
| GET | `/api/history` | `?day=YYYY-MM-DD` (default today): that day's completions; `&month=YYYY-MM` (default the day's month): daily counts |
| GET/POST | `/api/categories` | |
| PATCH/DELETE | `/api/categories/{id}` | delete unpins and uncategorises |
| POST | `/api/tasks` | `scheduled_today: true` schedules for the server's today |
| GET/PATCH/DELETE | `/api/tasks/{id}` | PATCH is partial; `null` clears a field |
| POST | `/api/tasks/{id}/complete`, `/reopen` | |
| POST | `/api/tasks/{id}/subtasks` | |
| PUT | `/api/tasks/{id}/subtasks/order` | `{"ids": [...]}` |
| PATCH/DELETE | `/api/subtasks/{id}` | |
| GET/POST | `/api/projects` | active by default; `?all=1` adds completed, `?closed=1` lists archived and completed |
| GET/PATCH/DELETE | `/api/projects/{id}` | delete keeps tasks as standalone; PATCH `archived: true` hides it and its open tasks |
