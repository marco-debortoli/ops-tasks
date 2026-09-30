<script lang="ts">
	import { api } from '$lib/api';
	import { addDays, short, taskNumber } from '$lib/format';
	import { act, app, categoryColor, openModal, taskPath, today } from '$lib/state.svelte';
	import Checkbox from './Checkbox.svelte';
	import Pri from './Pri.svelte';

	/** Mobile: one-line rows (priority, name, due) instead of the table. */
	let { compact = false }: { compact?: boolean } = $props();

	/** 'all', 'none' (uncategorised) or a category id. */
	let filter = $state<'all' | 'none' | number>('all');

	const queue = $derived(app.dash!.queue);

	const chips = $derived.by(() => {
		const counts = new Map<number | null, number>();
		for (const t of queue) counts.set(t.effective_category_id, (counts.get(t.effective_category_id) ?? 0) + 1);
		const cats = app.dash!.categories.filter((c) => counts.has(c.id)).map((c) => ({
			key: c.id as number | 'none',
			label: c.name,
			color: c.color,
			count: counts.get(c.id)!
		}));
		if (counts.has(null))
			cats.push({ key: 'none', label: 'none', color: 'var(--color-muted)', count: counts.get(null)! });
		return cats;
	});

	// Drop a filter whose category no longer has queued tasks.
	$effect(() => {
		if (filter !== 'all' && !chips.some((c) => c.key === filter)) filter = 'all';
	});

	const shown = $derived(
		filter === 'all'
			? queue
			: queue.filter((t) => (filter === 'none' ? t.effective_category_id == null : t.effective_category_id === filter))
	);

	const soon = $derived(addDays(today(), 7));
	const cols = 'grid-cols-[18px_44px_24px_minmax(0,1fr)_170px_44px_44px_34px]';
</script>

<section
	class="panel flex flex-col {compact ? 'px-2.5 pt-3.5 pb-1' : 'min-h-[320px] grow lg:min-h-0'}"
	aria-labelledby="queue-title"
>
	<h2 id="queue-title" class="panel-title {compact ? 'left-2' : ''}">
		queue <span class="font-normal text-muted">{queue.length}{compact ? '' : ' open, not for today'}</span>
	</h2>
	<div
		class="flex gap-1.5 pb-2 {compact ? '-mr-2.5 overflow-x-auto pr-2.5 [scrollbar-width:none] *:shrink-0' : 'flex-wrap'}"
		role="group"
		aria-label="Filter by category"
	>
		<button
			type="button"
			aria-pressed={filter === 'all'}
			class="btn-sm {compact ? 'h-8' : 'h-[26px]'} {filter === 'all' ? 'border-line-strong bg-line text-bright' : ''}"
			onclick={() => (filter = 'all')}>all {queue.length}</button
		>
		{#each chips as c (c.key)}
			<button
				type="button"
				aria-pressed={filter === c.key}
				class="btn-sm {compact ? 'h-8' : 'h-[26px]'} {filter === c.key ? 'border-line-strong bg-line' : ''}"
				style:color={c.color}
				onclick={() => (filter = c.key)}>{c.label} {c.count}</button
			>
		{/each}
	</div>

	{#if compact}
		{#each shown as t, i (t.id)}
			<button
				type="button"
				class="flex min-h-12 w-full items-center gap-2 text-left text-[13px] {i < shown.length - 1
					? 'border-b border-line-soft'
					: ''}"
				onclick={() => openModal({ kind: 'task', id: t.id })}
			>
				<span class="w-5 shrink-0"><Pri priority={t.priority} /></span>
				<span class="min-w-0 grow truncate text-bright">{t.name}</span>
				<span
					class="shrink-0 text-xs {t.due_date && t.due_date <= soon ? 'text-amber' : 'text-muted'}"
					title={t.due_date ? `due ${t.due_date}` : 'no due date'}>{short(t.due_date)}</span
				>
			</button>
		{:else}
			<p class="py-3 text-dim">queue is empty.</p>
		{/each}
	{:else}
		<div class="min-h-0 grow overflow-auto">
			<div class="min-w-[560px]">
				<div class="sticky top-0 grid {cols} gap-2 border-b border-line-soft bg-bg pb-1 text-xs text-dim">
					<span></span><span>ID</span><span>PRI</span><span>TASK</span><span>CATEGORY/PROJECT</span><span
						>SCHED</span
					><span>DUE</span><span>SUB</span>
				</div>
				{#each shown as t (t.id)}
					<div class="grid {cols} items-center gap-2 border-b border-line-soft py-1.5">
						<Checkbox checked={false} size={14} label="Complete {t.name}" onclick={() => act(() => api.completeTask(t.id)).catch(() => {})} />
						<button
							type="button"
							class="col-span-7 grid grid-cols-subgrid items-center text-left hover:bg-hover"
							onclick={() => openModal({ kind: 'task', id: t.id })}
						>
							<span class="text-dim">{taskNumber(t.id)}</span>
							<Pri priority={t.priority} />
							<span class="truncate text-bright">{t.name}</span>
							<span class="truncate" style:color={categoryColor(t.effective_category_id)}
								>{taskPath(t) || '—'}</span
							>
							<span class="text-muted" title={t.scheduled_date ?? ''}>{short(t.scheduled_date)}</span>
							<span
								class={t.due_date && t.due_date <= soon ? 'text-amber' : 'text-muted'}
								title={t.due_date ?? ''}>{short(t.due_date)}</span
							>
							<span class="text-muted">{t.subtasks_total ? `${t.subtasks_done}/${t.subtasks_total}` : '—'}</span>
						</button>
					</div>
				{:else}
					<p class="py-3 text-dim">queue is empty.</p>
				{/each}
			</div>
		</div>
	{/if}
</section>
