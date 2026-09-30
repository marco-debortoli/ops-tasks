<script lang="ts">
	import { addDays, isIsoDate } from '$lib/format';
	import { phone } from '$lib/media.svelte';
	import { today } from '$lib/state.svelte';

	let {
		id,
		value,
		onchange,
		quick = false,
		clearable = true,
		tone = 'var(--color-bright)',
		height,
		clearButton = true
	}: {
		id: string;
		value: string | null;
		onchange: (value: string | null) => void;
		/** Show today / tmrw / +1w / clear shortcuts. */
		quick?: boolean;
		clearable?: boolean;
		tone?: string;
		/** Defaults to 30px, or 44px on a phone. */
		height?: number;
		/** Show a clear button under the field (emptying the text also clears). */
		clearButton?: boolean;
	} = $props();

	let text = $state('');
	let invalid = $state(false);
	let picker: HTMLInputElement;

	$effect(() => {
		text = value ?? '';
		invalid = false;
	});

	function set(v: string | null) {
		invalid = false;
		text = v ?? '';
		if (v !== value) onchange(v);
	}

	function commit() {
		const t = text.trim();
		if (t === '') {
			if (clearable) set(null);
			else text = value ?? '';
		} else if (isIsoDate(t)) {
			set(t);
		} else {
			invalid = true;
		}
	}

	const shortcuts = $derived([
		{ label: 'today', value: today() },
		{ label: 'tmrw', value: addDays(today(), 1) },
		{ label: '+1w', value: addDays(today(), 7) }
	]);
</script>

<div class="flex flex-col gap-1">
	<div class="relative flex">
		<input
			{id}
			bind:value={text}
			placeholder="YYYY-MM-DD"
			autocomplete="off"
			spellcheck="false"
			aria-invalid={invalid}
			class="field min-w-0 grow rounded-r-none {invalid ? 'border-red!' : ''}"
			style:color={tone}
			style:height="{height ?? (phone.current ? 44 : 30)}px"
			onblur={commit}
			onkeydown={(e) => {
				if (e.key === 'Enter') commit();
			}}
		/>
		<button
			type="button"
			class="rounded-r-[3px] border border-l-0 border-line bg-raised px-2 text-dim hover:text-bright max-sm:px-3.5"
			aria-label="Pick a date"
			onclick={() => picker.showPicker()}>▾</button
		>
		<input
			bind:this={picker}
			type="date"
			tabindex="-1"
			aria-hidden="true"
			class="pointer-events-none absolute bottom-0 left-0 size-px opacity-0"
			value={value ?? ''}
			onchange={(e) => set(e.currentTarget.value || null)}
		/>
	</div>
	{#if invalid}
		<span class="text-xs text-red">use YYYY-MM-DD</span>
	{/if}
	{#if quick}
		<div class="flex flex-wrap gap-1">
			{#each shortcuts as s (s.label)}
				<button
					type="button"
					class="btn-sm {value === s.value ? 'border-line-strong bg-line text-bright' : ''}"
					onclick={() => set(s.value)}>{s.label}</button
				>
			{/each}
			{#if clearable && clearButton && value}
				<!-- `×` on a phone so the row fits a half-width field. -->
				<button type="button" class="btn-sm" aria-label="Clear" onclick={() => set(null)}
					><span class="max-sm:hidden">clear</span><span class="sm:hidden" aria-hidden="true">×</span></button
				>
			{/if}
		</div>
	{:else if clearable && clearButton && value}
		<div class="flex gap-1">
			<button type="button" class="btn-sm" onclick={() => set(null)}>clear</button>
		</div>
	{/if}
</div>
