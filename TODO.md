# TODO

Remaining work, based on [DESIGN_DECISIONS.md](DESIGN_DECISIONS.md) and the Terminal / ops
boards in the Task Manager Theme Explorations design canvas.

## MVP: next pass

### History
- [x] History modal (`OpsHistory` board): month calendar shaded by completions, the selected
      day's completed tasks beside it, buttons to step between days.
- [x] API endpoint for one day's completions and a month's per-day counts.
- [x] Enable the header history button (currently disabled).
- [x] Clicking a heatmap day opens history on that day (cells are plain spans today).

### Mobile
- [x] Mobile dashboard (`OpsMobile` board): stats strip, add input, Today, pinned
      (horizontal scroll), queue, below `lg`. The projects panel follows the queue so
      every project stays reachable; the header has history and categories icons.
- [x] Modals as bottom sheets below `sm` (`OpsMobileTask` board): drag the handle down
      to close; the task sheet has its own single-column layout.
- [ ] iOS Safari zooms into inputs under 16px on focus (ours are 13px). Either
      `maximum-scale=1` in the viewport meta or 16px inputs on phones.

## MVP: gaps in what's built

- [ ] Markdown preview for task descriptions (currently edit-only; sanitise the rendered HTML).
- [ ] A way to add a task straight to the queue (right now only via the Today input,
      then editing it, or from a project's modal).
- [ ] Drag-to-reorder subtasks (↑/↓ buttons exist).

## Phase 2 (in the mockups, not in the MVP)

- [ ] ⌘K command palette.
- [ ] Quick-add syntax: `p1 @project #category ^date !date`.
- [ ] Activity logs on tasks and projects.
- [ ] Next-7-days strip on the dashboard.
- [ ] Up/down changes on stats (vs previous week/month/year).
- [ ] Keyboard shortcuts beyond `esc` and `enter` (`x` complete, `s` reschedule, `1-3`
      priority, `t` add, `/` filter, `j/k` move, `h` history, `?` help).
- [ ] Duplicate task.
- [ ] Archive project.
- [ ] Project completion chart (completions over the last 6 weeks, pace, on-track estimate).

## Housekeeping

- [ ] `git init` and first commit.
- [ ] Integration tests for the store against a real Postgres (currently only pure-logic
      unit tests plus a manual API smoke script).
- [ ] Frontend tests (none yet).
- [ ] Wipe the demo data before real use: `docker compose down -v`.
- [ ] Set a real `POSTGRES_PASSWORD` in `.env` for the homelab deployment.
- [ ] Database backups for the homelab (e.g. scheduled `pg_dump`).
