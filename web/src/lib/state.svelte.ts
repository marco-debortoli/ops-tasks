import { api } from './api';
import type { Category, Dashboard, Task } from './types';

export type ModalEntry =
	| { kind: 'task'; id: number }
	| { kind: 'project'; id: number }
	| { kind: 'categories' }
	/** `day` defaults to today. */
	| { kind: 'history'; day?: string };

export const app = $state({
	dash: null as Dashboard | null,
	/** Last failed load or action, shown in the status bar. */
	error: '',
	syncedAt: null as Date | null,
	/** Open modals, bottom to top. */
	modals: [] as ModalEntry[]
});

export async function refresh() {
	try {
		app.dash = await api.dashboard();
		app.syncedAt = new Date();
		app.error = '';
	} catch (e) {
		app.error = e instanceof Error ? e.message : String(e);
	}
}

let refreshTimer: ReturnType<typeof setTimeout> | undefined;

/** Reload the dashboard shortly, coalescing bursts of changes. */
export function scheduleRefresh() {
	clearTimeout(refreshTimer);
	refreshTimer = setTimeout(refresh, 150);
}

/** Run an action, then refresh the dashboard; errors go to the status bar and are rethrown. */
export async function act<T>(fn: () => Promise<T>): Promise<T> {
	try {
		const result = await fn();
		scheduleRefresh();
		return result;
	} catch (e) {
		app.error = e instanceof Error ? e.message : String(e);
		throw e;
	}
}

export function openModal(m: ModalEntry) {
	app.modals.push(m);
}

export function closeModal(index: number) {
	app.modals.splice(index);
}

export function category(id: number | null | undefined): Category | undefined {
	if (id == null) return undefined;
	return app.dash?.categories.find((c) => c.id === id);
}

export function timezone(): string {
	return app.dash?.timezone ?? 'America/Vancouver';
}

export function today(): string {
	return app.dash?.today ?? new Date().toISOString().slice(0, 10);
}

/** Display path for a task: `home/deck-refinish` in a project, `home/` standalone. */
export function taskPath(t: Pick<Task, 'effective_category_id' | 'project_name'>): string {
	const cat = category(t.effective_category_id)?.name ?? '';
	if (!cat && !t.project_name) return '';
	return `${cat}/${t.project_name ?? ''}`;
}

export function categoryColor(id: number | null | undefined): string {
	return category(id)?.color ?? 'var(--color-muted)';
}
