<script lang="ts">
	import { clock } from '$lib/format';
	import { app, timezone } from '$lib/state.svelte';

	const d = $derived(app.dash);
	const todayCount = $derived(
		d ? d.today_tasks.overdue.length + d.today_tasks.due_today.length + d.today_tasks.scheduled.length : 0
	);
</script>

<!-- Desktop only, except to show an error on mobile. -->
<footer
	class="{app.error ? 'flex' : 'hidden lg:flex'} h-[26px] shrink-0 items-center justify-between gap-3 border-t border-line bg-raised px-3 text-[11.5px]"
>
	<div class="flex min-w-0 items-center gap-3.5">
		<span class="bg-blue px-1.5 py-px font-bold text-bg">NORMAL</span>
		{#if d}
			<span class="hidden truncate text-muted sm:inline">
				tasks · {todayCount} today · {d.queue.length} queued · {d.projects.filter((p) => p.pinned)
					.length} pinned · {d.projects.length} projects
			</span>
		{/if}
	</div>
	<div class="min-w-0 truncate" role="status">
		{#if app.error}
			<span class="text-red">E: {app.error}</span>
		{:else if app.syncedAt}
			<span class="text-muted"><span class="text-green">●</span> synced {clock(app.syncedAt, timezone())}</span>
		{/if}
	</div>
</footer>
