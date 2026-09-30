<script lang="ts">
	import { api } from '$lib/api';
	import { act } from '$lib/state.svelte';

	/** `large` is the 44px touch-sized input on the mobile dashboard. */
	let { large = false }: { large?: boolean } = $props();

	let draft = $state('');

	async function add(e: SubmitEvent) {
		e.preventDefault();
		const name = draft.trim();
		if (!name) return;
		draft = '';
		await act(() => api.createTask({ name, scheduled_today: true })).catch(() => (draft = name));
	}
</script>

<form
	onsubmit={add}
	class="flex shrink-0 items-center gap-2 rounded-[3px] border border-line-strong bg-raised {large
		? 'h-11 px-3'
		: 'h-[34px] px-2.5'}"
>
	<span class="font-bold text-green" aria-hidden="true">❯</span>
	<input
		bind:value={draft}
		aria-label="Add a task for today"
		placeholder={large ? 'add a task for today' : 'add a task for today, enter to save'}
		enterkeyhint="done"
		class="min-w-0 grow bg-transparent text-bright outline-none {large ? 'text-[13px]' : ''}"
	/>
</form>
