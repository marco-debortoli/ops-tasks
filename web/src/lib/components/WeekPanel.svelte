<script lang="ts">
	import { addDays, weekday } from '$lib/format';
	import { app, today } from '$lib/state.svelte';

	const tt = $derived(app.dash!.today_tasks);
	const open = $derived([...tt.overdue, ...tt.due_today, ...tt.scheduled, ...app.dash!.queue]);

	// Today counts what's in the Today panel: overdue tasks are due, missed scheduled ones roll
	// forward. A task counts once per day, as due if it's due that day, otherwise as scheduled.
	const days = $derived.by(() => {
		const t0 = today();
		return Array.from({ length: 7 }, (_, i) => {
			const date = addDays(t0, i);
			const on = (d: string | null) => d != null && (i === 0 ? d <= date : d === date);
			const due = open.filter((t) => on(t.due_date));
			const sched = open.filter((t) => !on(t.due_date) && on(t.scheduled_date));
			return {
				date,
				isToday: i === 0,
				n: due.length + sched.length,
				due: due.length,
				sched: sched.length,
				names: [...due, ...sched].map((t) => t.name)
			};
		});
	});
</script>

<section class="panel shrink-0" aria-labelledby="week-title">
	<h2 id="week-title" class="panel-title">next 7d</h2>
	<ol class="m-0 grid list-none grid-cols-7 gap-1.5 p-0">
		{#each days as d (d.date)}
			<li
				class="flex min-w-0 flex-col gap-0.5 rounded-[3px] border px-2 py-1.5 {d.isToday
					? 'border-amber bg-amber-deep'
					: 'border-line bg-raised'}"
				title={d.names.length ? `${d.date}\n${d.names.join('\n')}` : d.date}
			>
				<span class="truncate text-xs text-muted">{weekday(d.date)} {d.date.slice(8)}</span>
				<span
					class="text-[18px] leading-tight font-bold {d.n === 0 ? 'text-faint' : d.isToday ? 'text-amber' : 'text-bright'}"
					>{d.n}</span
				>
				<!-- Wraps between the counts, not inside one, when the column is narrow. -->
				<span class="flex flex-wrap gap-x-1.5 text-xs text-muted">
					{#if d.due}<span>{d.due} due</span>{/if}
					{#if d.sched}<span>{d.sched} sched</span>{/if}
					{#if !d.n}<span>clear</span>{/if}
				</span>
			</li>
		{/each}
	</ol>
</section>
