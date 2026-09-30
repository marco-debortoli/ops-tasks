<script lang="ts">
	import { daysLeft, percent, progressBar, short, statusColor, statusLabel } from '$lib/format';
	import { app, category, openModal, today } from '$lib/state.svelte';

	/** Mobile: one row of fixed-width cards that scrolls sideways. */
	let { scroll = false }: { scroll?: boolean } = $props();

	const pinned = $derived(
		app
			.dash!.projects.filter((p) => p.pinned)
			.sort((a, b) =>
				(category(a.category_id)?.name ?? '').localeCompare(category(b.category_id)?.name ?? '')
			)
	);
</script>

<section class="panel shrink-0 pb-3 {scroll ? 'pt-3.5 pr-0 pl-2.5' : ''}" aria-labelledby="pinned-title">
	<h2 id="pinned-title" class="panel-title {scroll ? 'left-2' : ''}">
		pinned <span class="font-normal text-muted">{scroll && pinned.length > 1 ? 'swipe' : '1 per category'}</span>
	</h2>
	{#if pinned.length}
		<div
			class={scroll
				? 'flex snap-x snap-mandatory scroll-pl-0 gap-2 overflow-x-auto pr-2.5 [scrollbar-width:none]'
				: 'grid grid-cols-1 gap-2.5 md:grid-cols-3'}
		>
			{#each pinned as p (p.id)}
				{@const cat = category(p.category_id)}
				{@const pct = percent(p.done_count, p.open_count + p.done_count)}
				<button
					type="button"
					class="flex min-w-0 flex-col gap-1.5 rounded-[3px] border border-line bg-raised p-2.5 text-left hover:border-line-strong {scroll
						? 'w-[230px] shrink-0 snap-start'
						: ''}"
					onclick={() => openModal({ kind: 'project', id: p.id })}
				>
					<span class="flex w-full justify-between gap-2 text-xs">
						<span style:color={cat?.color}>{cat?.name}</span>
						<span class="font-bold" style:color={statusColor[p.status]}>{statusLabel[p.status]}</span>
					</span>
					<span class="truncate font-bold text-bright">{p.name}</span>
					<span class="tracking-[-0.5px] text-green"
						>{progressBar(pct)} <span class="text-text">{pct}%</span></span
					>
					<span class="flex w-full justify-between gap-2 text-xs text-muted">
						<span>{p.open_count} open</span>
						<span
							>{#if p.due_date}due {short(p.due_date)} · {daysLeft(p.due_date, today())}{:else}no due
								date{/if}</span
						>
					</span>
					<span class="w-full truncate border-t border-dashed border-line pt-1.5 text-xs text-muted">
						next ❯ <span class="text-text">{p.next_task?.name ?? '—'}</span>
					</span>
				</button>
			{/each}
		</div>
	{:else}
		<p class="text-dim">no pinned projects. open a project and pin it; one per category.</p>
	{/if}
</section>
