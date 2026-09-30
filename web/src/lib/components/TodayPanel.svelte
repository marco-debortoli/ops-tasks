<script lang="ts">
	import { api } from '$lib/api';
	import { clock, daysBetween, short, subtaskBar } from '$lib/format';
	import { act, app, categoryColor, openModal, taskPath, timezone, today } from '$lib/state.svelte';
	import type { Task } from '$lib/types';
	import AddTask from './AddTask.svelte';
	import Checkbox from './Checkbox.svelte';
	import Pri from './Pri.svelte';

	const tt = $derived(app.dash!.today_tasks);
	const openCount = $derived(tt.overdue.length + tt.due_today.length + tt.scheduled.length);

	type Group = { key: string; label: string; color: string; tasks: Task[]; note: (t: Task) => [string, string] };

	const groups = $derived<Group[]>([
		{
			key: 'overdue',
			label: 'overdue',
			color: 'var(--color-red)',
			tasks: tt.overdue,
			note: (t) => [`due ${short(t.due_date)} · ${daysBetween(t.due_date!, today())}d late`, 'text-red']
		},
		{
			key: 'due',
			label: 'due today',
			color: 'var(--color-amber)',
			tasks: tt.due_today,
			note: () => ['due today', 'text-amber']
		},
		{
			key: 'scheduled',
			label: 'scheduled',
			color: 'var(--color-blue)',
			tasks: tt.scheduled,
			note: (t) => {
				const since = t.scheduled_date! < today() ? `since ${short(t.scheduled_date)} · ` : '';
				return [since + (t.due_date ? `due ${short(t.due_date)}` : 'no due date'), 'text-muted'];
			}
		},
		{
			key: 'completed',
			label: 'completed today',
			color: 'var(--color-muted)',
			tasks: tt.completed,
			note: (t) => [
				`completed ${t.completed_at ? clock(t.completed_at, timezone()) : ''} · ${t.completed_on}`,
				'text-dim'
			]
		}
	]);

	function toggle(t: Task) {
		act(() => (t.completed_on ? api.reopenTask(t.id) : api.completeTask(t.id))).catch(() => {});
	}
</script>

<section class="panel flex min-h-0 flex-col gap-1.5" aria-labelledby="today-title">
	<h2 id="today-title" class="panel-title">
		today <span class="font-normal text-muted">{openCount} open · {tt.completed.length} done</span>
	</h2>
	<AddTask />

	<div class="-mx-1 min-h-0 grow overflow-y-auto px-1">
		{#each groups as g (g.key)}
			{#if g.tasks.length}
				<div class="flex flex-col">
					<h3
						class="m-0 flex justify-between border-b border-line-soft pt-2 pb-1 text-xs font-bold tracking-[0.5px]"
					>
						<span style:color={g.color}>{g.label}</span><span class="text-dim">{g.tasks.length}</span>
					</h3>
					{#each g.tasks as t (t.id)}
						{@const [note, noteClass] = g.note(t)}
						{@const done = !!t.completed_on}
						<div class="flex items-start gap-2.5 border-b border-line-soft py-[7px]">
							<div class="mt-px">
								<Checkbox
									checked={done}
									label="{done ? 'Reopen' : 'Complete'} {t.name}"
									onclick={() => toggle(t)}
								/>
							</div>
							<button
								type="button"
								class="flex min-w-0 grow flex-col gap-[3px] text-left hover:bg-hover"
								onclick={() => openModal({ kind: 'task', id: t.id })}
							>
								<span class="flex gap-2">
									{#if !done && t.priority}<Pri priority={t.priority} />{/if}
									<span class={done ? 'text-dim line-through' : 'text-bright'}>{t.name}</span>
								</span>
								<span class="flex flex-wrap gap-x-2.5 text-xs text-muted">
									{#if taskPath(t)}
										<span style:color={categoryColor(t.effective_category_id)}>{taskPath(t)}</span>
									{/if}
									{#if t.subtasks_total && !done}<span>{subtaskBar(t.subtasks_done, t.subtasks_total)}</span>{/if}
									<span class={noteClass}>{note}</span>
								</span>
							</button>
						</div>
					{/each}
				</div>
			{/if}
		{/each}
		{#if openCount === 0 && tt.completed.length === 0}
			<p class="py-3 text-dim">nothing for today. add a task above, or schedule one from the queue.</p>
		{/if}
	</div>

	<div class="shrink-0 border-t border-line pt-2 text-xs leading-[1.7] text-muted">
		<span class="key">enter</span> open · <span class="key">esc</span> close · tasks scheduled
		earlier roll forward here until done
	</div>
</section>
