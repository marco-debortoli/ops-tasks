<script lang="ts">
	import { api } from '$lib/api';
	import { clock, daysBetween, short } from '$lib/format';
	import { act, app, categoryColor, openModal, taskPath, timezone, today } from '$lib/state.svelte';
	import type { Task } from '$lib/types';
	import Checkbox from './Checkbox.svelte';
	import Pri from './Pri.svelte';

	const tt = $derived(app.dash!.today_tasks);
	const openCount = $derived(tt.overdue.length + tt.due_today.length + tt.scheduled.length);

	// One list in the desktop panel's group order; the note carries the group instead of a heading.
	type Row = { task: Task; note: string; tone: string };
	const rows = $derived<Row[]>([
		...tt.overdue.map((t) => ({ task: t, note: `${daysBetween(t.due_date!, today())}d late`, tone: 'text-red' })),
		...tt.due_today.map((t) => ({ task: t, note: 'due today', tone: 'text-amber' })),
		...tt.scheduled.map((t) => ({
			task: t,
			note:
				(t.scheduled_date! < today() ? `since ${short(t.scheduled_date)}` : 'scheduled') +
				(t.due_date ? ` · due ${short(t.due_date)}` : ''),
			tone: 'text-blue'
		})),
		...tt.completed.map((t) => ({
			task: t,
			note: `done ${t.completed_at ? clock(t.completed_at, timezone()) : ''}`,
			tone: 'text-dim'
		}))
	]);

	function toggle(t: Task) {
		act(() => (t.completed_on ? api.reopenTask(t.id) : api.completeTask(t.id))).catch(() => {});
	}
</script>

<section class="panel px-2.5 pt-3.5 pb-1" aria-labelledby="m-today-title">
	<h2 id="m-today-title" class="panel-title left-2">
		today <span class="font-normal text-muted">{openCount} open{tt.completed.length ? ` · ${tt.completed.length} done` : ''}</span>
	</h2>
	{#each rows as { task: t, note, tone }, i (t.id)}
		{@const done = !!t.completed_on}
		<div class="flex min-h-[52px] items-center gap-1 {i < rows.length - 1 ? 'border-b border-line-soft' : ''}">
			<div class="-ml-2.5">
				<Checkbox touch checked={done} label="{done ? 'Reopen' : 'Complete'} {t.name}" onclick={() => toggle(t)} />
			</div>
			<button
				type="button"
				class="flex min-h-11 min-w-0 grow flex-col justify-center gap-0.5 text-left text-[13px]"
				onclick={() => openModal({ kind: 'task', id: t.id })}
			>
				<span class="flex min-w-0 gap-2">
					{#if !done && t.priority}<Pri priority={t.priority} />{/if}
					<span class="truncate {done ? 'text-dim line-through' : 'text-bright'}">{t.name}</span>
				</span>
				<span class="flex min-w-0 gap-2 text-xs">
					{#if taskPath(t)}
						<span class="truncate" style:color={categoryColor(t.effective_category_id)}>{taskPath(t)}</span>
					{/if}
					{#if t.subtasks_total && !done}<span class="shrink-0 text-muted">{t.subtasks_done}/{t.subtasks_total}</span>{/if}
					<span class="shrink-0 {tone}">{note}</span>
				</span>
			</button>
		</div>
	{:else}
		<p class="py-3 text-dim">nothing for today. add a task above.</p>
	{/each}
</section>
