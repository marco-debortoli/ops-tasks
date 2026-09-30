<script lang="ts">
	import { api } from '$lib/api';
	import { percent, short, statusColor, statusLabel } from '$lib/format';
	import { act, app, category, openModal } from '$lib/state.svelte';
	import type { ProjectStatus } from '$lib/types';

	let adding = $state(false);
	let draft = $state('');

	const projects = $derived(app.dash!.projects);
	const order: ProjectStatus[] = ['in_progress', 'waiting', 'todo'];
	const groups = $derived(
		order
			.map((status) => ({ status, items: projects.filter((p) => p.status === status) }))
			.filter((g) => g.items.length)
	);

	async function create(e: SubmitEvent) {
		e.preventDefault();
		const name = draft.trim();
		if (!name) return;
		try {
			const p = await act(() => api.createProject({ name }));
			draft = '';
			adding = false;
			openModal({ kind: 'project', id: p.id });
		} catch {
			// shown in the status bar
		}
	}
</script>

<section class="panel flex min-h-[240px] grow flex-col lg:min-h-0" aria-labelledby="projects-title">
	<h2 id="projects-title" class="panel-title">
		projects <span class="font-normal text-muted">{projects.length} active</span>
	</h2>
	<button
		type="button"
		class="absolute -top-[11px] right-2.5 bg-bg px-1.5 text-xs text-green hover:text-bright"
		aria-expanded={adding}
		onclick={() => (adding = !adding)}>+ project</button
	>

	{#if adding}
		<form
			onsubmit={create}
			class="mb-1.5 flex h-8 shrink-0 items-center gap-2 rounded-[3px] border border-line-strong bg-raised px-2.5"
		>
			<span class="font-bold text-green" aria-hidden="true">❯</span>
			<!-- svelte-ignore a11y_autofocus -->
			<input
				bind:value={draft}
				autofocus
				aria-label="New project name"
				placeholder="project name, enter to create"
				class="min-w-0 grow bg-transparent text-bright outline-none"
				onkeydown={(e) => {
					if (e.key === 'Escape') {
						adding = false;
						draft = '';
					}
				}}
			/>
		</form>
	{/if}

	<div class="-mx-1 min-h-0 grow overflow-y-auto px-1">
		{#each groups as g (g.status)}
			<h3
				class="m-0 border-b border-line-soft pt-1.5 pb-[3px] text-xs font-bold"
				style:color={statusColor[g.status]}
			>
				{statusLabel[g.status]} <span class="font-normal text-dim">{g.items.length}</span>
			</h3>
			{#each g.items as p (p.id)}
				{@const cat = category(p.category_id)}
				<button
					type="button"
					class="grid w-full grid-cols-[12px_minmax(0,1fr)_40px_44px] items-center gap-1.5 border-b border-line-soft py-[5px] text-left hover:bg-hover"
					onclick={() => openModal({ kind: 'project', id: p.id })}
				>
					<span class="text-amber" aria-label={p.pinned ? 'pinned' : undefined}>{p.pinned ? '◆' : ''}</span>
					<span class="flex min-w-0 flex-col">
						<span class="truncate text-bright">{p.name}</span>
						<span class="text-xs" style:color={cat?.color ?? 'var(--color-dim)'}>{cat?.name ?? 'none'}</span>
					</span>
					<span class="text-right text-text">{percent(p.done_count, p.open_count + p.done_count)}%</span>
					<span class="text-right text-xs text-muted" title={p.due_date ?? ''}>{short(p.due_date)}</span>
				</button>
			{/each}
		{:else}
			<p class="py-2 text-dim">no active projects.</p>
		{/each}
	</div>
</section>
