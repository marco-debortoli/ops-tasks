import type {
	Category,
	Dashboard,
	History,
	Project,
	ProjectDetail,
	ProjectPatch,
	Subtask,
	TaskDetail,
	TaskPatch
} from './types';

export class ApiError extends Error {
	constructor(
		public status: number,
		message: string
	) {
		super(message);
	}
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const res = await fetch('/api' + path, {
		method,
		headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
		body: body === undefined ? undefined : JSON.stringify(body)
	});
	if (res.status === 204) return undefined as T;
	const data = await res.json().catch(() => ({}));
	if (!res.ok) throw new ApiError(res.status, data.error ?? res.statusText);
	return data as T;
}

export const api = {
	dashboard: () => request<Dashboard>('GET', '/dashboard'),
	/** Tasks completed on `day` (default today) and daily counts for `month` (default day's). */
	history: (day?: string, month?: string) => {
		const q = new URLSearchParams();
		if (day) q.set('day', day);
		if (month) q.set('month', month);
		return request<History>('GET', '/history' + (q.size ? '?' + q : ''));
	},

	createCategory: (name: string, color: string) =>
		request<Category>('POST', '/categories', { name, color }),
	updateCategory: (id: number, patch: Partial<Pick<Category, 'name' | 'color'>>) =>
		request<Category>('PATCH', `/categories/${id}`, patch),
	deleteCategory: (id: number) => request<void>('DELETE', `/categories/${id}`),

	createTask: (
		input: TaskPatch & { name: string; scheduled_today?: boolean }
	) => request<TaskDetail>('POST', '/tasks', input),
	getTask: (id: number) => request<TaskDetail>('GET', `/tasks/${id}`),
	updateTask: (id: number, patch: TaskPatch) => request<TaskDetail>('PATCH', `/tasks/${id}`, patch),
	deleteTask: (id: number) => request<void>('DELETE', `/tasks/${id}`),
	completeTask: (id: number) => request<TaskDetail>('POST', `/tasks/${id}/complete`),
	reopenTask: (id: number) => request<TaskDetail>('POST', `/tasks/${id}/reopen`),

	createSubtask: (taskId: number, name: string) =>
		request<Subtask>('POST', `/tasks/${taskId}/subtasks`, { name }),
	updateSubtask: (id: number, patch: Partial<Pick<Subtask, 'name' | 'done'>>) =>
		request<Subtask>('PATCH', `/subtasks/${id}`, patch),
	deleteSubtask: (id: number) => request<void>('DELETE', `/subtasks/${id}`),
	reorderSubtasks: (taskId: number, ids: number[]) =>
		request<Subtask[]>('PUT', `/tasks/${taskId}/subtasks/order`, { ids }),

	listProjects: (all = false) => request<Project[]>('GET', '/projects' + (all ? '?all=1' : '')),
	createProject: (input: ProjectPatch & { name: string }) =>
		request<ProjectDetail>('POST', '/projects', input),
	getProject: (id: number) => request<ProjectDetail>('GET', `/projects/${id}`),
	updateProject: (id: number, patch: ProjectPatch) =>
		request<ProjectDetail>('PATCH', `/projects/${id}`, patch),
	deleteProject: (id: number) => request<void>('DELETE', `/projects/${id}`)
};
