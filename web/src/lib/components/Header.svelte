<script lang="ts">
	import { onMount } from 'svelte';
	import { headerClock, short, weekday } from '$lib/format';
	import { openModal, timezone, today } from '$lib/state.svelte';

	let now = $state(new Date());
	onMount(() => {
		const t = setInterval(() => (now = new Date()), 10_000);
		return () => clearInterval(t);
	});
</script>

<header
	class="flex h-14 shrink-0 items-center justify-between gap-4 border-b border-line bg-raised pr-2 pl-3 lg:h-11 lg:px-3"
>
	<h1 class="m-0 flex items-center gap-2.5 text-base">
		<span class="bg-green px-2 py-1 font-bold text-bg">ops</span>
		<span class="font-bold text-bright">tasks</span>
		<span class="text-xs font-normal text-dim lg:hidden">{weekday(today())} {short(today())}</span>
	</h1>

	<!-- Desktop: text buttons and the clock. -->
	<div class="hidden items-center gap-2 lg:flex">
		<span class="mr-1.5 text-blue">{headerClock(now, timezone())}</span>
		<button type="button" class="btn h-[30px]" onclick={() => openModal({ kind: 'categories' })}>
			categories
		</button>
		<button type="button" class="btn h-[30px]" onclick={() => openModal({ kind: 'history' })}>
			history
		</button>
	</div>

	<!-- Mobile: 44px icon buttons. -->
	<div class="flex gap-0.5 lg:hidden">
		<button
			type="button"
			class="flex size-11 items-center justify-center text-text"
			aria-label="History"
			onclick={() => openModal({ kind: 'history' })}
		>
			<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true">
				<path d="M3 12a9 9 0 1 0 3-6.7" /><path d="M3 4v5h5" /><path d="M12 8v4l3 2" />
			</svg>
		</button>
		<button
			type="button"
			class="flex size-11 items-center justify-center text-text"
			aria-label="Categories"
			onclick={() => openModal({ kind: 'categories' })}
		>
			<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true">
				<path d="M3 12V4h8l10 10-8 8L3 12z" /><circle cx="7.5" cy="8.5" r="1.5" />
			</svg>
		</button>
	</div>
</header>
