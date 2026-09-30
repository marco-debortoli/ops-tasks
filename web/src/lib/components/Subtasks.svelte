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

	function moveTo(from: number, to: number) {
		if (from === to || to < 0 || to >= items.length) return;
		const [s] = items.splice(from, 1);
		items.splice(to, 0, s);
		const ids = items.map((s) => s.id);
		save(() => api.reorderSubtasks(taskId, ids));
	}

	// Drag to reorder, with pointer events so mouse and touch share one path. Rows stay put in
	// `items` while dragging: the dragged row follows the pointer and the rows it passes shift
	// over by its height, then the order is committed on release.
	let rows = $state<HTMLElement[]>([]);
	let drag = $state<{ from: number; to: number; dy: number; height: number } | null>(null);
	/** Row midpoints at drag start; rows don't move in layout while dragging, only by transform. */
	let mids: number[] = [];
	let startY = 0;

	function dragStart(e: PointerEvent, i: number) {
		if (e.button !== 0 || editing !== null) return;
		e.preventDefault();
		(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
		mids = rows.slice(0, items.length).map((r) => {
			const b = r.getBoundingClientRect();
			return b.top + b.height / 2;
		});
		startY = e.clientY;
		drag = { from: i, to: i, dy: 0, height: rows[i].getBoundingClientRect().height };
	}

	function dragMove(e: PointerEvent) {
		if (!drag) return;
		drag.dy = e.clientY - startY;
		const y = mids[drag.from] + drag.dy;
		// New index = how many of the other rows now sit above the dragged row's centre.
		drag.to = mids.filter((m, j) => j !== drag!.from && m < y).length;
	}

	function dragEnd() {
		if (!drag) return;
		const { from, to } = drag;
		drag = null;
		moveTo(from, to);
	}

	/** Vertical offset for row `i` while a drag is in progress. */
	function shift(i: number): number {
		if (!drag) return 0;
		const { from, to, dy, height } = drag;
		if (i === from) return dy;
		if (from < to && i > from && i <= to) return -height;
		if (to < from && i >= to && i < from) return height;
		return 0;
	}
</script>

<div class="flex flex-col gap-0.5">
	<div class="flex items-baseline justify-between border-b border-line-soft pb-1.5">
		<h3 class="m-0 text-xs font-bold text-green">subtasks</h3>
		<span class="text-xs text-muted">{subtaskBar(done, items.length)}</span>
	</div>

	{#each items as s, i (s.id)}
		<div
			bind:this={rows[i]}
			class="group relative flex min-h-[34px] items-center gap-2.5 border-b border-line-soft pr-1.5 max-sm:min-h-11 {drag?.from ===
			i
				? 'z-10 border-line-strong bg-raised shadow-[0_4px_12px_rgb(0_0_0/0.5)]'
				: drag
					? 'transition-transform duration-150'
					: 'hover:bg-hover focus-within:bg-hover'}"
			style:transform={drag ? `translateY(${shift(i)}px)` : undefined}
		>
			<!-- Drag handle. A button so touch targeting prefers it over the checkbox beside it; the ↑/↓
			     buttons are the keyboard route, so it stays out of the tab order. -->
			<button
				type="button"
				tabindex="-1"
				aria-hidden="true"
				title="drag to reorder"
				class="-mr-2.5 flex w-4 shrink-0 cursor-grab touch-none items-center justify-center self-stretch text-faint select-none hover:text-soft max-sm:mr-0 max-sm:w-8 {drag?.from ===
				i
					? 'cursor-grabbing text-soft'
					: ''}"
				onpointerdown={(e) => dragStart(e, i)}
				onpointermove={dragMove}
				onpointerup={dragEnd}
				onpointercancel={() => (drag = null)}>⠿</button
			>
			<div class="max-sm:-mr-3">
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
						onclick={() => moveTo(i, i - 1)}>↑</button
					>
					<button
						type="button"
						class="btn-sm"
						aria-label="Move {s.name} down"
						disabled={i === items.length - 1}
						onclick={() => moveTo(i, i + 1)}>↓</button
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
