// Mirrors the JSON returned by the Go API (backend/internal/store).

export type Priority = 1 | 2 | 3;
export type ProjectStatus = 'todo' | 'in_progress' | 'waiting' | 'complete';

export interface Category {
	id: number;
	name: string;
	color: string;
}

export interface Task {
	id: number;
	name: string;
	description: string;
	priority: Priority | null;
	scheduled_date: string | null;
	due_date: string | null;
	project_id: number | null;
	project_name: string | null;
	/** The task's own category; only standalone tasks have one. */
	category_id: number | null;
	/** What to display: the project's category, or the task's own. */
	effective_category_id: number | null;
	completed_on: string | null;
	completed_at: string | null;
	subtasks_done: number;
	subtasks_total: number;
	created_at: string;
	updated_at: string;
}

export interface Subtask {
	id: number;
	task_id: number;
	name: string;
	done: boolean;
	position: number;
}

export interface TaskDetail extends Task {
	subtasks: Subtask[];
}

export interface TaskRef {
	id: number;
	name: string;
	priority: Priority | null;
}

export interface Project {
	id: number;
	name: string;
	category_id: number | null;
	status: ProjectStatus;
	due_date: string | null;
	notes: string;
	pinned: boolean;
	/** Archived projects, and their open tasks, are hidden from the dashboard. */
	archived_at: string | null;
	open_count: number;
	done_count: number;
	next_task: TaskRef | null;
	created_at: string;
	updated_at: string;
}

export interface ProjectDetail extends Project {
	open_tasks: Task[];
	completed_tasks: Task[];
}

export interface DayCount {
	date: string;
	count: number;
}

export interface Stats {
	week: number;
	month: number;
	year: number;
	/** The same stretch of the previous week, month and year. */
	prev_week: number;
	prev_month: number;
	prev_year: number;
	streak: number;
	heatmap: DayCount[];
	by_category: { category_id: number | null; count: number }[];
}

export interface Dashboard {
	today: string;
	timezone: string;
	categories: Category[];
	today_tasks: {
		overdue: Task[];
		due_today: Task[];
		scheduled: Task[];
		completed: Task[];
	};
	queue: Task[];
	projects: Project[];
	stats: Stats;
}

export interface History {
	day: string;
	/** Completed on `day`, in the order they were done. */
	tasks: Task[];
	/** `YYYY-MM` */
	month: string;
	/** Every day of `month`. */
	days: DayCount[];
}

/** Fields accepted by PATCH /api/tasks/{id}; null clears a field. */
export type TaskPatch = Partial<{
	name: string;
	description: string;
	priority: Priority | null;
	scheduled_date: string | null;
	due_date: string | null;
	project_id: number | null;
	category_id: number | null;
	completed_on: string;
}>;

export type ProjectPatch = Partial<{
	name: string;
	category_id: number | null;
	status: ProjectStatus;
	due_date: string | null;
	notes: string;
	pinned: boolean;
	archived: boolean;
}>;
