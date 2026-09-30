import type { Priority, ProjectStatus } from './types';

// Dates are ISO `YYYY-MM-DD` strings throughout; "today" comes from the server.

const DAY_MS = 86_400_000;

function utc(iso: string): number {
	const [y, m, d] = iso.split('-').map(Number);
	return Date.UTC(y, m - 1, d);
}

/** Whole days from a to b (positive when b is later). */
export function daysBetween(a: string, b: string): number {
	return Math.round((utc(b) - utc(a)) / DAY_MS);
}

export function addDays(iso: string, n: number): string {
	return new Date(utc(iso) + n * DAY_MS).toISOString().slice(0, 10);
}

const WEEKDAYS = ['sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat'];
const MONTHS = [
	'january', 'february', 'march', 'april', 'may', 'june',
	'july', 'august', 'september', 'october', 'november', 'december'
];

/** `thu` */
export function weekday(iso: string): string {
	return WEEKDAYS[new Date(utc(iso)).getUTCDay()];
}

/** Monday-based weekday index, 0-6. */
export function weekdayIndex(iso: string): number {
	return (new Date(utc(iso)).getUTCDay() + 6) % 7;
}

/** `YYYY-MM` of a date. */
export const monthOf = (iso: string) => iso.slice(0, 7);

/** `YYYY-MM` shifted by n months. */
export function addMonths(month: string, n: number): string {
	const [y, m] = month.split('-').map(Number);
	return new Date(Date.UTC(y, m - 1 + n, 1)).toISOString().slice(0, 7);
}

/** `september 2026`, or just `sep` when short. */
export function monthName(month: string, shortName = false): string {
	const name = MONTHS[Number(month.slice(5, 7)) - 1];
	return shortName ? name.slice(0, 3) : `${name} ${month.slice(0, 4)}`;
}

/** `today`, `yesterday`, `4 days ago`, `in 2 days`. */
export function relativeDay(iso: string, today: string): string {
	const n = daysBetween(iso, today);
	if (n === 0) return 'today';
	if (n === 1) return 'yesterday';
	return n > 0 ? `${n} days ago` : `in ${-n} days`;
}

/** Short date: `MM/DD`, or an em dash when unset. */
export function short(iso: string | null | undefined): string {
	if (!iso) return '—';
	return `${iso.slice(5, 7)}/${iso.slice(8, 10)}`;
}

export function isIsoDate(s: string): boolean {
	return /^\d{4}-\d{2}-\d{2}$/.test(s) && !Number.isNaN(utc(s)) && addDays(s, 0) === s;
}

/** Time of day `HH:MM` for a timestamp, in the app's time zone. */
export function clock(ts: string | Date, timeZone: string): string {
	return new Intl.DateTimeFormat('en-GB', {
		timeZone,
		hour: '2-digit',
		minute: '2-digit',
		hourCycle: 'h23'
	}).format(new Date(ts));
}

/** Calendar date `YYYY-MM-DD` of a timestamp, in the app's time zone. */
export function dateIn(ts: string | Date, timeZone: string): string {
	return new Intl.DateTimeFormat('en-CA', { timeZone }).format(new Date(ts));
}

/** Header clock: `wed 2026-09-30 14:02` in the app's time zone. */
export function headerClock(now: Date, timeZone: string): string {
	const parts = Object.fromEntries(
		new Intl.DateTimeFormat('en-CA', {
			timeZone,
			weekday: 'short',
			year: 'numeric',
			month: '2-digit',
			day: '2-digit',
			hour: '2-digit',
			minute: '2-digit',
			hourCycle: 'h23'
		})
			.formatToParts(now)
			.map((p) => [p.type, p.value])
	);
	return `${parts.weekday.toLowerCase()} ${parts.year}-${parts.month}-${parts.day} ${parts.hour}:${parts.minute}`;
}

/** Relative deadline: `17d left`, `due today`, `3d late`. */
export function daysLeft(due: string, today: string): string {
	const n = daysBetween(today, due);
	if (n === 0) return 'due today';
	return n > 0 ? `${n}d left` : `${-n}d late`;
}

/** `[██████░░░░]` */
export function progressBar(pct: number, width = 10): string {
	const n = Math.round((pct / 100) * width);
	return '[' + '█'.repeat(n) + '░'.repeat(width - n) + ']';
}

/** `▰▰▱▱▱ 2/5`, scaled down to at most 10 blocks. */
export function subtaskBar(done: number, total: number): string {
	if (!total) return '';
	const width = Math.min(total, 10);
	const filled = Math.round((done / total) * width);
	return '▰'.repeat(filled) + '▱'.repeat(width - filled) + ` ${done}/${total}`;
}

export function percent(done: number, total: number): number {
	return total ? Math.round((done / total) * 100) : 0;
}

export function taskNumber(id: number): string {
	return '#' + String(id).padStart(4, '0');
}

export const priorityLabel = (p: Priority | null) => (p ? `P${p}` : '');

export const priorityColor: Record<Priority, string> = {
	1: 'var(--color-red)',
	2: 'var(--color-amber)',
	3: 'var(--color-muted)'
};

export const statusLabel: Record<ProjectStatus, string> = {
	todo: 'TODO',
	in_progress: 'IN PROGRESS',
	waiting: 'WAITING',
	complete: 'COMPLETE'
};

export const statusColor: Record<ProjectStatus, string> = {
	todo: 'var(--color-muted)',
	in_progress: 'var(--color-green)',
	waiting: 'var(--color-amber)',
	complete: 'var(--color-blue)'
};
