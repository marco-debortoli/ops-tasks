<script lang="ts">
	import { app } from '$lib/state.svelte';

	const stats = $derived(app.dash!.stats);
	const cells = $derived([
		{ label: 'week', value: String(stats.week), tone: 'text-bright' },
		{ label: 'month', value: String(stats.month), tone: 'text-bright' },
		{ label: 'year', value: String(stats.year), tone: 'text-bright' },
		{ label: 'streak', value: `${stats.streak}d`, tone: 'text-green' }
	]);
</script>

<!-- Mobile stand-in for the stats panel: completion counts only. -->
<section class="grid shrink-0 grid-cols-4 rounded-[3px] border border-line" aria-label="Completed tasks">
	{#each cells as c, i (c.label)}
		<div class="px-2 py-1.5 {i < 3 ? 'border-r border-line' : ''}">
			<div class="text-[17px] font-bold {c.tone}">{c.value}</div>
			<div class="text-xs text-muted">{c.label}</div>
		</div>
	{/each}
</section>
