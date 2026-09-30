<script lang="ts">
	import type { Snippet } from 'svelte';
	import { clock } from '$lib/format';
	import { timezone } from '$lib/state.svelte';

	let {
		status,
		onclose,
		children
	}: {
		status: { savedAt: Date | null; error: string };
		onclose: () => void;
		children: Snippet;
	} = $props();
</script>

<!-- Phone: sits under the sheet's grab handle, with a 44px close button. -->
<div
	class="flex h-10 shrink-0 items-center justify-between gap-3 border-b border-line bg-raised px-3 max-sm:h-11 max-sm:border-line-soft max-sm:bg-modal max-sm:pr-1 max-sm:pl-4"
>
	<div class="flex min-w-0 items-center gap-2.5 max-sm:gap-2 max-sm:text-sm">
		{@render children()}
	</div>
	<div class="flex shrink-0 items-center gap-2.5">
		{#if status.error}
			<span class="text-xs text-red" role="alert">E: {status.error}</span>
		{:else if status.savedAt}
			<span class="text-xs text-dim max-sm:hidden">:w saved {clock(status.savedAt, timezone())}</span>
		{/if}
		<button type="button" class="btn-sm h-[26px] text-text max-sm:hidden" aria-label="Close" onclick={onclose}>
			esc
		</button>
		<button
			type="button"
			class="flex size-11 items-center justify-center text-text sm:hidden"
			aria-label="Close"
			onclick={onclose}
		>
			<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true">
				<path d="M6 6l12 12" /><path d="M18 6L6 18" />
			</svg>
		</button>
	</div>
</div>
