<script lang="ts">
	import { untrack } from 'svelte';
	import { api } from '$lib/api';
	import { autosave } from '$lib/autosave.svelte';
	import {
		addDays,
		addMonths,
		clock,
		daysBetween,
		dateIn,
		monthName,
		monthOf,
		relativeDay,
		short,
		weekday,
		weekdayIndex
	} from '$lib/format';
	import {
		app,
		category,
		categoryColor,
		openModal,
		scheduleRefresh,
		taskPath,
		timezone,
		today
	} from '$lib/state.svelte';
	import type { History, Task } from '$lib/types';
	import DateField from './DateField.svelte';
	import Modal from './Modal.svelte';
	import ModalHeader from './ModalHeader.svelte';
	import Pri from './Pri.svelte';

	let { day: initialDay, onclose }: { day?: string; onclose: () => void } = $props();

	// The selected day, and the month the calendar shows (which can be browsed away from it).
	const startDay = untrack(() => initialDay) ?? today();
	let day = $state(startDay);
	let month = $state(monthOf(startDay));
	let history = $state<History | null>(null);
	let loadError = $state('');
	let editing = $state<number | null>(null);

	const saver = autosave<{ id: number; on: string }>(async ({ id, on }) => {
		await api.updateTask(id, { completed_on: on });
		editing = null;
		scheduleRefresh();
	});

	// Ignore responses that arrive after a newer request, when stepping through days quickly.
	let requestSeq = 0;
	async function load(d: string, m: string) {
		const seq = ++requestSeq;
		try {
			const h = await api.history(d, m);
			if (seq !== requestSeq) return;
			history = h;
			loadError = '';
		} catch (e) {
			if (seq === requestSeq) loadError = e instanceof Error ? e.message : String(e);
		}
	}

	// Reload on navigation, and whenever the dashboard reloads (a task may have changed).
	$effect(() => {
		void app.syncedAt;
		load(day, month);
	});

	function select(d: string) {
		day = d;
		month = monthOf(d);
		editing = null;
	}

	const isToday = $derived(day === today());
	const ago = $derived(daysBetween(day, today()));
	const prevDay = $derived(addDays(day, -1));
	const nextDay = $derived(addDays(day, 1));
	const dayLabel = (d: string) => `${weekday(d)} ${Number(d.slice(8))}`;

	// Calendar: Monday-first weeks, blank cells before the 1st.
	const cells = $derived.by(() => {
		if (!history || history.month !== month) return [];
		const lead = weekdayIndex(history.days[0].date);
		return [...Array(lead).fill(null), ...history.days] as (History['days'][number] | null)[];
	});

	// The calendar uses four shades; the brightest is kept for the dashboard heatmap.
	const shade = (n: number) => `var(--color-heat-${n <= 0 ? 0 : n === 1 ? 1 : n === 2 ? 2 : 3})`;

	const monthStats = $derived.by(() => {
		if (!history || history.month !== month) return null;
		const total = history.days.reduce((sum, d) => sum + d.count, 0);
		const best = Math.max(0, ...history.days.map((d) => d.count));
		// Average over the days that have happened so far.
		const elapsed = history.days.filter((d) => d.date <= today()).length;
		return { total, best, perDay: elapsed ? (total / elapsed).toFixed(1) : '—' };
	});

	const byCategory = $derived.by(() => {
		const counts = new Map<number | null, number>();
		for (const t of history?.tasks ?? []) {
			counts.set(t.effective_category_id, (counts.get(t.effective_category_id) ?? 0) + 1);
		}
		return [...counts]
			.sort((a, b) => b[1] - a[1])
			.map(([id, n]) => ({
				key: id ?? 'none',
				label: category(id)?.name ?? 'none',
				color: categoryColor(id),
				n
			}));
	});

	const projectsTouched = $derived([
		...new Set((history?.tasks ?? []).flatMap((t) => (t.project_name ? [t.project_name] : [])))
	]);

	function meta(t: Task): string {
		const parts: string[] = [];
		if (t.due_date) {
			const late = daysBetween(t.due_date, day);
			parts.push(late > 0 ? `was due ${short(t.due_date)} · ${late}d late` : `was due ${short(t.due_date)}`);
		}
		if (t.subtasks_total) parts.push(`${t.subtasks_done}/${t.subtasks_total} subtasks`);
		// The completion date can be edited, so say when it was actually ticked off.
		if (t.completed_at) {
			const actual = dateIn(t.completed_at, timezone());
			if (actual !== day) parts.push(`ticked off ${short(actual)}`);
		}
		return parts.join(' · ');
	}
</script>

<Modal labelledby="hist-title" width={1000} height={720} {onclose}>
	<ModalHeader status={saver.status} {onclose}>
		<h2 id="hist-title" class="m-0 text-base font-bold text-green">history</h2>
		<span class="hidden text-dim sm:inline">completed tasks by date</span>
		{#if ago >= 0}
			<span class="bg-amber px-1.5 py-px text-xs font-bold text-bg">{ago ? `HEAD~${ago}` : 'HEAD'}</span>
		{/if}
	</ModalHeader>

	{#if loadError && !history}
		<p class="p-4 text-red">E: {loadError}</p>
	{:else}
		<div
			class="grid min-h-0 grow grid-cols-1 content-start overflow-y-auto md:content-normal md:grid-cols-[360px_minmax(0,1fr)] md:overflow-hidden"
		>
			<!-- Month calendar -->
			<div
				class="flex flex-col gap-3 border-b border-line-soft bg-side px-[18px] py-4 md:border-r md:border-b-0"
			>
				<div class="flex items-center justify-between">
					<button
						type="button"
						class="btn h-7 px-2 text-soft"
						aria-label="Previous month"
						onclick={() => (month = addMonths(month, -1))}
						>‹ {monthName(addMonths(month, -1), true)}</button
					>
					<span class="font-bold text-bright">{monthName(month)}</span>
					<button
						type="button"
						class="btn h-7 px-2 text-soft"
						aria-label="Next month"
						disabled={month >= monthOf(today())}
						onclick={() => (month = addMonths(month, 1))}
						>{monthName(addMonths(month, 1), true)} ›</button
					>
				</div>
				<div class="grid grid-cols-7 gap-1 text-center text-xs text-dim" aria-hidden="true">
					{#each ['mo', 'tu', 'we', 'th', 'fr', 'sa', 'su'] as d (d)}<span>{d}</span>{/each}
				</div>
				<div class="grid grid-cols-7 gap-1">
					{#each cells as c, i (c?.date ?? i)}
						{#if !c}
							<span class="h-10"></span>
						{:else}
							{@const future = c.date > today()}
							{@const selected = c.date === day}
							<button
								type="button"
								disabled={future}
								aria-label="{c.date}, {c.count} completed"
								aria-pressed={selected}
								class="flex h-10 flex-col items-center justify-center gap-px rounded-[3px] p-0 hover:brightness-125 disabled:hover:brightness-100
									{selected ? 'border-2 border-amber' : c.date === today() ? 'border border-green' : 'border border-transparent'}"
								style:background={future ? 'transparent' : shade(c.count)}
								onclick={() => select(c.date)}
							>
								<span class="text-xs {future ? 'text-faint' : 'text-text'} {selected ? 'font-bold' : ''}"
									>{Number(c.date.slice(8))}</span
								>
								{#if !future}
									<span class="text-xs {c.count >= 3 ? 'text-bright' : 'text-soft'}">{c.count}</span>
								{/if}
							</button>
						{/if}
					{/each}
				</div>
				<div class="flex items-center gap-1.5 text-xs text-dim" aria-hidden="true">
					less
					{#each [0, 1, 2, 3] as n (n)}<span class="size-2.5" style:background={shade(n)}></span>{/each}
					more
				</div>
				{#if monthStats}
					<div class="grid grid-cols-3 gap-2 border-t border-dashed border-line pt-2.5">
						{#each [{ n: monthStats.total, label: monthName(month).split(' ')[0] }, { n: monthStats.perDay, label: 'per day' }, { n: monthStats.best, label: 'best day' }] as s (s.label)}
							<div>
								<div class="text-[22px] font-bold text-bright">{s.n}</div>
								<div class="text-xs text-muted">{s.label}</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Selected day -->
			<div class="flex min-h-0 flex-col gap-2.5 px-5 py-4">
				<div class="flex flex-wrap items-center justify-between gap-2">
					<div class="flex flex-col gap-0.5">
						<span class="text-lg font-bold text-bright">{weekday(day)} {day}</span>
						<span class="text-xs text-muted">
							{history?.day === day ? history.tasks.length : '…'} completed · {relativeDay(day, today())}
						</span>
					</div>
					<div class="flex gap-1.5">
						<button type="button" class="btn h-[30px] px-2.5" onclick={() => select(prevDay)}
							>‹ {dayLabel(prevDay)}</button
						>
						<button
							type="button"
							class="btn h-[30px] px-2.5"
							disabled={nextDay > today()}
							onclick={() => select(nextDay)}>{dayLabel(nextDay)} ›</button
						>
						<button
							type="button"
							class="btn h-[30px] border-amber-edge px-2.5 text-amber hover:border-amber"
							disabled={isToday && month === monthOf(day)}
							onclick={() => select(today())}>back to today</button
						>
					</div>
				</div>

				<div
					class="hidden grid-cols-[50px_26px_minmax(0,1fr)_180px_110px] gap-2.5 border-b border-line-soft pb-1 text-xs text-dim sm:grid"
				>
					<span>TIME</span><span>PRI</span><span>TASK</span><span>CATEGORY/PROJECT</span><span></span>
				</div>
				<div class="-mx-1 min-h-0 grow overflow-y-auto px-1">
					{#if history?.day === day}
						{#each history.tasks as t (t.id)}
							<div
								class="grid grid-cols-[50px_26px_minmax(0,1fr)_auto] items-center gap-2.5 border-b border-line-soft py-2 sm:grid-cols-[50px_26px_minmax(0,1fr)_180px_110px]"
							>
								<span class="text-muted">{t.completed_at ? clock(t.completed_at, timezone()) : '—'}</span>
								<Pri priority={t.priority} />
								<button
									type="button"
									class="flex min-w-0 flex-col gap-0.5 text-left hover:bg-hover"
									onclick={() => openModal({ kind: 'task', id: t.id })}
								>
									<span class="max-w-full truncate text-bright">{t.name}</span>
									{#if meta(t)}<span class="max-w-full truncate text-xs text-dim">{meta(t)}</span>{/if}
								</button>
								<span class="hidden truncate sm:inline" style:color={categoryColor(t.effective_category_id)}
									>{taskPath(t)}</span
								>
								<span class="flex justify-end">
									{#if editing === t.id}
										<button type="button" class="btn-sm h-[26px]" onclick={() => (editing = null)}>cancel</button>
									{:else}
										<button type="button" class="btn-sm h-[26px]" onclick={() => (editing = t.id)}
											>change date</button
										>
									{/if}
								</span>
								{#if editing === t.id}
									<div class="col-span-full flex items-start gap-2.5 pb-1 sm:col-start-3">
										<label for="hist-date-{t.id}" class="label pt-2">completed on</label>
										<div class="w-[170px]">
											<DateField
												id="hist-date-{t.id}"
												value={t.completed_on}
												clearable={false}
												onchange={(v) => v && saver.save({ id: t.id, on: v })}
											/>
										</div>
									</div>
								{/if}
							</div>
						{:else}
							<p class="py-3 text-dim">nothing completed on {weekday(day)} {day}.</p>
						{/each}
					{:else if loadError}
						<p class="py-3 text-red">E: {loadError}</p>
					{:else}
						<p class="py-3 text-dim">loading…</p>
					{/if}
				</div>

				{#if history?.day === day && history.tasks.length}
					<div
						class="flex shrink-0 flex-wrap gap-x-4 gap-y-1 border-t border-dashed border-line pt-2.5 text-xs text-muted"
					>
						<span>by category</span>
						{#each byCategory as k (k.key)}
							<span style:color={k.color}>{k.label} {k.n}</span>
						{/each}
						{#if projectsTouched.length}
							<span class="ml-auto">projects touched: {projectsTouched.join(', ')}</span>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</Modal>
