import { addDays, daysBetween, short } from './format';

export const PACE_WEEKS = 6;

export interface Pace {
	/** Completions per rolling 7-day window, oldest first; the last one ends today. */
	weeks: { from: string; to: string; count: number }[];
	/** Completions per week over the whole span. */
	perWeek: number;
	/** When the open tasks will be done at this pace; null when nothing is open or the pace is zero. */
	estimate: string | null;
	/** Whether the estimate beats the due date; null when there's nothing to compare. */
	onTrack: boolean | null;
	/** `pace: 1.5/wk · 4 open · on track for 10/15` */
	summary: string;
}

/** Completion pace over the last six weeks, from the dates tasks were completed on. */
export function projectPace(
	completedOn: string[],
	open: number,
	today: string,
	due: string | null
): Pace {
	const weeks = Array.from({ length: PACE_WEEKS }, (_, i) => {
		const to = addDays(today, -7 * (PACE_WEEKS - 1 - i));
		const from = addDays(to, -6);
		return { from, to, count: completedOn.filter((d) => d >= from && d <= to).length };
	});
	const perWeek = weeks.reduce((n, w) => n + w.count, 0) / PACE_WEEKS;
	const base = { weeks, perWeek, estimate: null, onTrack: null };
	const head = `pace: ${perWeek.toFixed(1)}/wk · ${open} open`;

	if (open === 0) return { ...base, summary: `${head} · nothing left` };
	if (perWeek === 0) {
		return { ...base, onTrack: due ? false : null, summary: `${head} · no recent completions` };
	}
	const estimate = addDays(today, Math.ceil((open / perWeek) * 7));
	if (!due) return { ...base, estimate, summary: `${head} · est. done ${short(estimate)}` };
	if (estimate <= due) {
		return { ...base, estimate, onTrack: true, summary: `${head} · on track for ${short(due)}` };
	}
	return {
		...base,
		estimate,
		onTrack: false,
		summary: `${head} · est. ${short(estimate)}, ${daysBetween(due, estimate)}d past due`
	};
}
