<script lang="ts">
	import { api } from '$lib/api';
	import { subtaskBar } from '$lib/format';
	import type { Subtask } from '$lib/types';
	import { phone } from '$lib/media.svelte';
	import Checkbox from './Checkbox.svelte';

	let {
		taskId,
		subtasks,
		save
	}: {
		taskId: number;
		subtasks: Subtask[];
		/** Runs a change through the modal's autosave (status line + dashboard refresh). */
		save: (fn: () => Promise<unknown>) => Promise<void>;
	} = $props();

	let items = $state<Subtask[]>([]);
	let draft = $state('');
	let editing = $state<number | null>(null);
	let editText = $state('');
	/** Row whose actions are showing; touch screens have no hover, so tapping the name toggles them. */
	let active = $state<number | null>(null);

	$effect(() => {
		items = subtasks.map((s) => ({ ...s }));
	});

	const done = $derived(items.filter((s) => s.done).length);

	async function add(e: SubmitEvent) {
		e.preventDefault();
		const name = draft.trim();
		if (!name) return;
		draft = '';
		await save(async () => {
			items.push(await api.createSubtask(taskId, name));
		});
	}

	function toggle(s: Subtask) {
		s.done = !s.done;
		save(() => api.updateSubtask(s.id, { done: s.done }));
	}

	function startEdit(s: Subtask) {
		editing = s.id;
		editText = s.name;
	}

	function commitEdit(s: Subtask) {
		const name = editText.trim();
		editing = null;
		if (!name || name === s.name) return;
		s.name = name;
		save(() => api.updateSubtask(s.id, { name }));
	}

	function remove(s: Subtask) {
		items = items.filter((x) => x.id !== s.id);
		save(() => api.deleteSubtask(s.id));
	}

	function move(i: number, by: number) {
		const j = i + by;
		if (j < 0 || j >= items.length) return;
		[items[i], items[j]] = [items[j], items[i]];
		const ids = items.map((s) => s.id);
		save(() => api.reorderSubtasks(taskId, ids));
	}
</script>

<div class="flex flex-col gap-0.5">
	<div class="flex items-baseline justify-between border-b border-line-soft pb-1.5">
		<h3 class="m-0 text-xs font-bold text-green">subtasks</h3>
		<span class="text-xs text-muted">{subtaskBar(done, items.length)}</span>
	</div>

	{#each items as s, i (s.id)}
		<div
			class="group flex min-h-[34px] items-center max-sm:min-h-11 gap-2.5 border-b border-line-soft px-1.5 hover:bg-hover focus-within:bg-hover"
		>
			<div class="max-sm:-mx-3">
				<Checkbox touch={phone.current} checked={s.done} label="Toggle {s.name}" onclick={() => toggle(s)} />
			</div>
			{#if editing === s.id}
				<!-- svelte-ignore a11y_autofocus -->
				<input
					bind:value={editText}
					aria-label="Subtask name"
					autofocus
					class="field h-[26px] min-w-0 grow"
					onblur={() => commitEdit(s)}
					onkeydown={(e) => {
						if (e.key === 'Enter') commitEdit(s);
						if (e.key === 'Escape') {
							e.preventDefault();
							e.stopPropagation();
							editing = null;
						}
					}}
				/>
			{:else}
				<button
					type="button"
					class="grow self-stretch text-left {s.done ? 'text-dim line-through' : 'text-bright'}"
					aria-expanded={active === s.id}
					onclick={() => (active = active === s.id ? null : s.id)}>{s.name}</button
				>
				<span class="{active === s.id ? 'flex' : 'hidden'} gap-1.5 group-focus-within:flex group-hover:flex">
					<button
						type="button"
						class="btn-sm"
						aria-label="Move {s.name} up"
						disabled={i === 0}
						onclick={() => move(i, -1)}>↑</button
					>
					<button
						type="button"
						class="btn-sm"
						aria-label="Move {s.name} down"
						disabled={i === items.length - 1}
						onclick={() => move(i, 1)}>↓</button
					>
					<button type="button" class="btn-sm" onclick={() => startEdit(s)}>edit</button>
					<button type="button" class="btn-sm btn-danger" onclick={() => remove(s)}>del</button>
				</span>
			{/if}
		</div>
	{/each}

	<form
		onsubmit={add}
		class="mt-1.5 flex h-8 items-center gap-2 max-sm:h-11 rounded-[3px] border border-dashed border-line-strong bg-raised px-2.5"
	>
		<span class="text-green" aria-hidden="true">+</span>
		<input
			bind:value={draft}
			aria-label="Add subtask"
			placeholder="add subtask, enter to save"
			class="min-w-0 grow bg-transparent text-bright outline-none"
		/>
	</form>
</div>
