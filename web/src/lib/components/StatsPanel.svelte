<script lang="ts">
	import { change } from '$lib/format';
	import { app, category, openModal, today } from '$lib/state.svelte';

	const stats = $derived(app.dash!.stats);
	const counts = $derived([
		{ n: stats.week, prev: stats.prev_week, label: 'week' },
		{ n: stats.month, prev: stats.prev_month, label: 'month' },
		{ n: stats.year, prev: stats.prev_year, label: 'year' }
	]);

	function level(n: number): number {
		if (n <= 0) return 0;
		if (n === 1) return 1;
		if (n === 2) return 2;
		if (n <= 4) return 3;
		return 4;
	}

	const byCategory = $derived.by(() => {
		const max = Math.max(1, ...stats.by_category.map((k) => k.count));
		return stats.by_category.map((k) => {
			const c = category(k.category_id);
			return {
				key: k.category_id ?? 'none',
				label: c?.name ?? 'none',
				color: c?.color ?? 'var(--color-muted)',
				count: k.count,
				width: Math.round((k.count / max) * 100)
			};
		});
	});
</script>

<section class="panel flex shrink-0 flex-col gap-2.5 pb-3" aria-labelledby="stats-title">
	<h2 id="stats-title" class="panel-title">
		stats <span class="font-normal text-muted">completed</span>
	</h2>
	<div class="grid grid-cols-3 gap-2">
		{#each counts as s (s.label)}
			{@const c = change(s.n, s.prev)}
			<div title="{s.prev} by this point last {s.label}">
				<div class="text-[26px] leading-[1.1] font-bold text-bright">{s.n}</div>
				<div class="text-xs text-muted">
					{s.label} <span style:color={c.color}>{c.text}</span>
					<span class="sr-only">vs {s.prev} by this point last {s.label}</span>
				</div>
			</div>
		{/each}
	</div>

	<div
		class="grid grid-flow-col grid-rows-[repeat(7,9px)] auto-cols-[9px] gap-[2px]"
		role="group"
		aria-label="Completions per day over the last 26 weeks"
	>
		{#each stats.heatmap as d (d.date)}
			{#if d.date > today()}
				<span></span>
			{:else}
				<!-- Out of the tab order: 182 stops is too many, and history steps through days itself. -->
				<button
					type="button"
					tabindex="-1"
					class="rounded-[1px] p-0 hover:outline hover:outline-soft"
					class:outline={d.date === today()}
					class:outline-amber={d.date === today()}
					style:background="var(--color-heat-{level(d.count)})"
					title="{d.date} · {d.count} completed"
					aria-label="{d.date}, {d.count} completed"
					onclick={() => openModal({ kind: 'history', day: d.date })}
				></button>
			{/if}
		{/each}
	</div>
	<div class="flex justify-between text-xs text-muted">
		<span>26 weeks · mon–sun</span>
		<span>streak <span class="text-green">{stats.streak}d</span></span>
	</div>

	<div class="flex flex-col gap-1 border-t border-dashed border-line pt-2">
		<span class="text-xs text-dim">this year by category</span>
		{#each byCategory as k (k.key)}
			<div class="grid grid-cols-[60px_minmax(0,1fr)_34px] items-center gap-2 text-xs">
				<span class="truncate" style:color={k.color}>{k.label}</span>
				<span class="h-2 bg-[#1A2220]"
					><span class="block h-full" style:width="{k.width}%" style:background={k.color}></span></span
				>
				<span class="text-right text-text">{k.count}</span>
			</div>
		{:else}
			<span class="text-xs text-dim">nothing completed yet this year</span>
		{/each}
	</div>
</section>
