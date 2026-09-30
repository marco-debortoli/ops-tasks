# ops · tasks: design decisions

Summary of the decisions made during design.

## Scope

- A personal task manager only. Finance and adventure-tracking apps, multi-app
  navigation, and a theme system were considered and dropped to ship faster.
- Single user, no authentication. Runs in Docker on a homelab, reached over VPN.

## Visual style

- Terminal/ops look, fixed (not switchable): original phosphor-green and amber palette
  on near-black, JetBrains Mono throughout.
- Lowercase interface text; panels with their titles inset into the top border.
- Deliberately crowded dashboard. Everything opens in a modal (bottom sheet on mobile)
  instead of navigating to another page.
- Short dates `MM/DD`; full dates ISO 8601 `YYYY-MM-DD`.

## Dashboard layout

- Header: static "ops tasks" label, date and time, history button.
- Left column: Today. Middle: pinned projects, then the queue (backlog).
  Right: stats, then the projects list.
- Mobile stacks: stats strip, add input, Today, pinned (horizontal scroll), queue.

## Tasks

- Fields: name, markdown description, priority (P1, P2, P3, or none), scheduled date,
  due date, optional project, subtasks.
- Scheduled = when I plan to work on it. Due = deadline.
- Tasks are open or completed. No task statuses.
- Subtasks are checklist items: add, edit, delete, reorder, check off. Completing a task
  does not touch its subtasks.
- Modals autosave; no save/cancel buttons.

## Today panel

- Four groups, in order: overdue, due today, scheduled, completed today.
- An open task scheduled for an earlier day rolls forward into Today ("since MM/DD")
  instead of dropping into the queue.
- Typing a name in the Today input creates a task scheduled for today.
- The queue holds every other open task, sorted by priority then due date, with filter
  chips by category.

## Projects

- Statuses: TODO, IN PROGRESS, WAITING, COMPLETE. Show the full names, not the
  RUN/WAIT abbreviations in the mockups.
- Fields: name, optional category, optional due date, optional notes.
- The projects panel lists all unfinished projects grouped by status.

## Pinned (priority) projects

- One pinned project per category. Pinning another project in that category replaces it.
- Pinning requires a category. Changing a pinned project's category, or completing it,
  unpins it.
- A pinned card shows progress, open task count, days left, and the next task
  (highest-priority open task).

## Categories

- Simple labels with a name and a colour.
- A task in a project inherits the project's category and can't override it.
  A standalone task can have its own category or none.
- Moving a task out of its project copies the project's category onto the task.
  Deleting a project keeps its tasks as standalone tasks.
- Tasks display as a path: `home/deck-refinish` (in a project), `home/` (standalone).
- A simple category manager (add, rename, recolour, delete) is needed but wasn't mocked up.
- A "+ project" button in the projects panel header is needed but wasn't mocked up.

## Completion and history

- Completing a task records the date and the actual time. The date defaults to today
  and is editable afterwards; history and stats use that date.
- History modal: month calendar shaded by completions, the selected day's completed
  tasks beside it, and buttons to step between days.
- Clicking a day in the dashboard heatmap opens history on that day.

## Stats

- Tasks completed this week (Monday start), this month, and this year.
- Daily streak, 26-week heatmap, and completions this year by category.

## Phase 2 (in the mockups, not in the MVP)

- ⌘K command palette; quick-add syntax (`p1 @project #category ^date !date`).
- Activity logs; next-7-days strip; up/down changes on stats.
- Keyboard shortcuts beyond `esc` and `enter`.
- Duplicate task, archive project, project completion chart.

## Technical

- SvelteKit single-page app built as static files, served by a Go backend; Postgres.
- "Today" is calculated on the server in `America/Vancouver`.
