<script lang="ts">
	import { api } from '$lib/api';
	import { autosave } from '$lib/autosave.svelte';
	import { app, scheduleRefresh } from '$lib/state.svelte';
	import type { Category } from '$lib/types';
	import Modal from './Modal.svelte';
	import ModalHeader from './ModalHeader.svelte';

	let { onclose }: { onclose: () => void } = $props();

	const presets = [
		'#7FD1C7',
		'#D6A3DE',
		'#86B8E3',
		'#8FD694',
		'#E3A77F',
		'#D9D27A',
		'#EF7A6D',
		'#A7B4AE'
	];

	let newName = $state('');
	let newColor = $state(presets[0]);
	let confirmDelete = $state<number | null>(null);

	// Every change here is a single request; autosave just provides the status line.
	const saver = autosave<() => Promise<unknown>>(async (fn) => {
		await fn();
		scheduleRefresh();
	});
	const run = (fn: () => Promise<unknown>) => saver.save(fn);

	async function add(e: SubmitEvent) {
		e.preventDefault();
		const name = newName.trim();
		if (!name) return;
		await run(() => api.createCategory(name, newColor));
		if (!saver.status.error) {
			newName = '';
			newColor = presets[(presets.indexOf(newColor) + 1) % presets.length];
		}
	}

	function rename(c: Category, input: HTMLInputElement) {
		const name = input.value.trim();
		if (!name) input.value = c.name;
		else if (name !== c.name) run(() => api.updateCategory(c.id, { name }));
	}

	// How much a delete would touch, from what the dashboard knows about.
	function usage(c: Category): string {
		const d = app.dash;
		if (!d) return '';
		const projects = d.projects.filter((p) => p.category_id === c.id).length;
		const all = [
			...d.queue,
			...d.today_tasks.overdue,
			...d.today_tasks.due_today,
			...d.today_tasks.scheduled
		];
		const tasks = all.filter((t) => t.effective_category_id === c.id).length;
		return `${projects} projects · ${tasks} open tasks`;
	}
</script>

<Modal labelledby="cat-title" width={620} height={560} onclose={onclose}>
	<ModalHeader status={saver.status} {onclose}>
		<h2 id="cat-title" class="m-0 text-base font-bold text-green">categories</h2>
		<span class="text-dim">{app.dash?.categories.length ?? 0}</span>
	</ModalHeader>

	<div class="flex min-h-0 grow flex-col gap-1 overflow-y-auto px-[18px] py-3.5">
		<div class="grid grid-cols-[34px_minmax(0,1fr)_150px_auto] gap-2.5 border-b border-line-soft pb-1 text-xs text-dim">
			<span>COL</span><span>NAME</span><span>USED BY</span><span></span>
		</div>
		{#each app.dash?.categories ?? [] as c (c.id)}
			<div class="grid grid-cols-[34px_minmax(0,1fr)_150px_auto] items-center gap-2.5 border-b border-line-soft py-1.5">
				<input
					type="color"
					value={c.color}
					aria-label="Colour of {c.name}"
					class="h-6 w-[34px] cursor-pointer rounded-[2px] border border-line bg-transparent p-0"
					onchange={(e) => {
						const color = e.currentTarget.value;
						run(() => api.updateCategory(c.id, { color }));
					}}
				/>
				<input
					value={c.name}
					aria-label="Name of {c.name}"
					class="field h-7"
					style:color={c.color}
					onblur={(e) => rename(c, e.currentTarget)}
					onkeydown={(e) => e.key === 'Enter' && e.currentTarget.blur()}
				/>
				<span class="text-xs text-muted">{usage(c)}</span>
				{#if confirmDelete === c.id}
					<span class="flex gap-1">
						<button
							type="button"
							class="btn-sm btn-danger"
							onclick={() => {
								confirmDelete = null;
								run(() => api.deleteCategory(c.id));
							}}>confirm</button
						>
						<button type="button" class="btn-sm" onclick={() => (confirmDelete = null)}>keep</button>
					</span>
				{:else}
					<button type="button" class="btn-sm btn-danger" onclick={() => (confirmDelete = c.id)}
						>del</button
					>
				{/if}
			</div>
		{:else}
			<p class="py-2 text-dim">no categories yet</p>
		{/each}

		<form onsubmit={add} class="mt-3 flex flex-col gap-2">
			<div
				class="flex h-8 items-center gap-2 rounded-[3px] border border-dashed border-line-strong bg-raised px-2.5"
			>
				<span class="text-green" aria-hidden="true">+</span>
				<input
					bind:value={newName}
					aria-label="New category name"
					placeholder="new category, enter to save"
					class="min-w-0 grow bg-transparent outline-none"
					style:color={newColor}
				/>
			</div>
			<div class="flex flex-wrap items-center gap-1.5" role="group" aria-label="Colour">
				{#each presets as color (color)}
					<button
						type="button"
						aria-label="Colour {color}"
						aria-pressed={newColor === color}
						class="size-6 rounded-[2px] border-2 {newColor === color ? 'border-bright' : 'border-transparent'}"
						style:background={color}
						onclick={() => (newColor = color)}
					></button>
				{/each}
				<input
					type="color"
					bind:value={newColor}
					aria-label="Custom colour"
					class="h-6 w-8 cursor-pointer rounded-[2px] border border-line bg-transparent p-0"
				/>
			</div>
		</form>
		<p class="mt-auto pt-3 text-xs text-dim">
			deleting a category leaves its projects and tasks uncategorised and unpins its pinned project.
		</p>
	</div>
</Modal>
