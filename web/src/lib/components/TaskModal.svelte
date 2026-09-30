<script lang="ts">
	import { onDestroy, tick, untrack } from 'svelte';
	import { api } from '$lib/api';
	import { autosave } from '$lib/autosave.svelte';
	import { clock, dateIn, short, taskNumber } from '$lib/format';
	import { renderMarkdown } from '$lib/markdown';
	import { phone } from '$lib/media.svelte';
	import {
		app,
		category,
		categoryColor,
		openModal,
		scheduleRefresh,
		taskPath,
		timezone
	} from '$lib/state.svelte';
	import type { Priority, Project, TaskDetail, TaskPatch } from '$lib/types';
	import DateField from './DateField.svelte';
	import Modal from './Modal.svelte';
	import ModalHeader from './ModalHeader.svelte';
	import Subtasks from './Subtasks.svelte';

	let { id, onclose }: { id: number; onclose: () => void } = $props();

	let task = $state<TaskDetail | null>(null);
	let projects = $state<Project[]>([]);
	let loadError = $state('');
	let name = $state('');
	let description = $state('');
	let confirmDelete = $state(false);
	/** Show the description rendered, or the textarea. Starts on preview when there's something to show. */
	let previewing = $state(false);

	const saver = autosave<TaskPatch>(async (patch) => {
		task = await api.updateTask(id, patch);
		scheduleRefresh();
	});

	async function load() {
		try {
			const [t, ps] = await Promise.all([api.getTask(id), api.listProjects()]);
			if (!task) {
				name = t.name;
				description = t.description;
				previewing = !!t.description.trim();
			}
			task = t;
			projects = ps;
		} catch (e) {
			loadError = e instanceof Error ? e.message : String(e);
		}
	}

	// Load now, and again whenever the dashboard reloads (another modal may have changed this task).
	$effect(() => {
		void app.syncedAt;
		untrack(load);
	});

	function close() {
		saver.flush();
		onclose();
	}
	onDestroy(saver.flush);

	/** Run a change that isn't a task PATCH, with the same status line and refresh. */
	async function run(fn: () => Promise<unknown>) {
		await saver.flush();
		try {
			await fn();
			saver.status.savedAt = new Date();
			saver.status.error = '';
			scheduleRefresh();
		} catch (e) {
			saver.status.error = e instanceof Error ? e.message : String(e);
		}
	}

	async function setCompleted(done: boolean) {
		await run(async () => {
			task = await (done ? api.completeTask(id) : api.reopenTask(id));
		});
	}

	async function remove() {
		await run(() => api.deleteTask(id));
		if (!saver.status.error) onclose();
	}

	async function editDescription() {
		previewing = false;
		await tick();
		document.getElementById(`task-desc-${id}`)?.focus();
	}

	const priorities: { value: Priority | null; label: string; on: string }[] = [
		{ value: 1, label: 'P1', on: 'bg-red-deep text-red font-bold' },
		{ value: 2, label: 'P2', on: 'bg-amber-deep text-amber font-bold' },
		{ value: 3, label: 'P3', on: 'bg-line text-bright font-bold' },
		{ value: null, label: 'none', on: 'bg-line text-bright' }
	];

	// The project picker lists unfinished projects, plus the current one if it's complete.
	const projectOptions = $derived.by(() => {
		if (!task?.project_id || projects.some((p) => p.id === task!.project_id)) return projects;
		return [...projects, { id: task.project_id, name: task.project_name ?? '?' } as Project];
	});
</script>

<Modal labelledby="task-name-{id}" width={860} height={760} onclose={close}>
	<ModalHeader status={saver.status} onclose={close}>
		<span class="font-bold text-green">task</span>
		<span class="text-dim">{taskNumber(id)}</span>
		{#if task && taskPath(task)}
			<span class="text-dim max-sm:hidden">in</span>
			{#if task.project_id}
				<button
					type="button"
					class="truncate hover:underline"
					style:color={categoryColor(task.effective_category_id)}
					onclick={() => openModal({ kind: 'project', id: task!.project_id! })}
					>{taskPath(task)}</button
				>
			{:else}
				<span class="truncate" style:color={categoryColor(task.effective_category_id)}>{taskPath(task)}</span>
			{/if}
		{/if}
		{#if task?.completed_on}
			<span class="text-green max-sm:hidden">✓ completed</span>
		{/if}
	</ModalHeader>

	{#if loadError}
		<p class="p-4 text-red">E: {loadError}</p>
	{:else if task && phone.current}
		<!-- Phone: one scrolling column, completion up top, actions in a fixed footer. -->
		<div class="flex min-h-0 grow flex-col gap-3.5 overflow-y-auto px-4 py-3.5">
			<div class="flex flex-col gap-1">
				<label for="task-name-{id}" class="label">name</label>
				{@render nameInput('text-[17px]')}
			</div>

			{#if task.completed_on}
				<div class="flex flex-col gap-1.5 rounded-[3px] border border-heat-2 bg-[#111c14] px-3 py-2.5">
					<span class="flex items-center justify-between">
						<span class="font-bold text-green">✓ completed</span>
						<span class="text-xs text-muted">exited 0</span>
					</span>
					<label for="task-done-{id}" class="text-xs text-soft">completed on</label>
					<DateField
						id="task-done-{id}"
						value={task.completed_on}
						clearable={false}
						onchange={(v) => v && saver.save({ completed_on: v })}
					/>
					{#if task.completed_at}
						<span class="text-xs leading-normal text-muted">
							Marked done {short(dateIn(task.completed_at, timezone()))}
							{clock(task.completed_at, timezone())}. Change this if you finished it on a different day;
							history uses this date.
						</span>
					{/if}
				</div>
			{/if}

			{@render priorityField()}
			<div class="grid grid-cols-2 gap-2.5">
				{@render projectField()}
				{@render categoryField()}
			</div>
			<div class="grid grid-cols-2 gap-2.5">
				{@render scheduledField()}
				{@render dueField()}
			</div>
			{@render descriptionField(3)}
			<Subtasks taskId={id} subtasks={task.subtasks} save={run} />
			<div class="flex gap-2 pt-1">{@render deleteButtons()}</div>
		</div>

		<div
			class="grid shrink-0 grid-cols-2 gap-2.5 border-t border-line bg-raised px-4 pt-3 pb-[calc(20px+env(safe-area-inset-bottom))]"
		>
			{#if task.completed_on}
				<button
					type="button"
					class="btn h-12 justify-center border-amber-edge text-amber hover:border-amber"
					onclick={() => setCompleted(false)}>reopen</button
				>
			{:else}
				<button type="button" class="btn btn-primary h-12 justify-center" onclick={() => setCompleted(true)}
					>mark complete</button
				>
			{/if}
			<button
				type="button"
				class="btn h-12 justify-center border-line-strong bg-line font-bold text-bright"
				onclick={close}>close</button
			>
		</div>
	{:else if task}
		<div class="border-b border-line-soft px-[18px] pt-4 pb-2.5">
			<label for="task-name-{id}" class="label mb-1 block">name</label>
			{@render nameInput('text-xl')}
		</div>

		<div class="grid min-h-0 grow grid-cols-[minmax(0,1fr)_280px]">
			<div class="flex min-h-0 flex-col gap-3.5 overflow-y-auto border-r border-line-soft px-[18px] py-3.5">
				{@render descriptionField(6)}
				<Subtasks taskId={id} subtasks={task.subtasks} save={run} />
			</div>

			<div class="flex flex-col gap-3 bg-side px-4 py-3.5">
				{@render priorityField()}
				{@render projectField()}
				{@render categoryField()}
				{@render scheduledField()}
				{@render dueField()}

				<div class="flex flex-col gap-1.5">
					<label for="task-done-{id}" class="label">
						completed
						{#if task.completed_at}
							<span class="text-faint">· at {clock(task.completed_at, timezone())}</span>
						{/if}
					</label>
					{#if task.completed_on}
						<DateField
							id="task-done-{id}"
							value={task.completed_on}
							clearable={false}
							onchange={(v) => v && saver.save({ completed_on: v })}
						/>
					{:else}
						<input
							id="task-done-{id}"
							disabled
							placeholder="set when marked complete"
							class="h-[30px] rounded-[3px] border border-dashed border-line bg-transparent px-2.5"
						/>
					{/if}
				</div>
			</div>
		</div>

		<div
			class="flex h-[52px] shrink-0 items-center justify-between gap-2 border-t border-line bg-raised px-3.5"
		>
			<div class="flex items-center gap-2">{@render deleteButtons()}</div>
			<div class="flex items-center gap-2.5">
				<span class="text-xs text-dim"><span class="key">esc</span> close</span>
				{#if task.completed_on}
					<button type="button" class="btn" onclick={() => setCompleted(false)}>reopen</button>
				{:else}
					<button type="button" class="btn btn-primary" onclick={() => setCompleted(true)}
						>mark complete</button
					>
				{/if}
			</div>
		</div>
	{:else}
		<p class="p-4 text-dim">loading…</p>
	{/if}
</Modal>

<!-- Fields shared by the desktop and phone layouts; only rendered once `task` has loaded. -->

{#snippet nameInput(size: string)}
	<input
		id="task-name-{id}"
		bind:value={name}
		class="field h-10 w-full border-line-strong font-bold max-sm:h-11 {size} {name.trim() ? '' : 'border-red!'}"
		enterkeyhint="done"
		oninput={() => name.trim() && saver.queue({ name })}
		onkeydown={(e) => e.key === 'Enter' && saver.flush()}
		onblur={saver.flush}
	/>
{/snippet}

{#snippet descriptionField(rows: number)}
	<div class="flex flex-col gap-1.5">
		<div class="flex items-center justify-between">
			<label for="task-desc-{id}" class="label">description <span class="text-faint">· markdown</span></label>
			<div class="flex" role="group" aria-label="Description view">
				<button
					type="button"
					aria-pressed={!previewing}
					class="btn-sm rounded-r-none {previewing ? '' : 'border-line-strong bg-line text-bright'}"
					onclick={editDescription}>edit</button
				>
				<button
					type="button"
					aria-pressed={previewing}
					class="btn-sm -ml-px rounded-l-none {previewing ? 'border-line-strong bg-line text-bright' : ''}"
					onclick={() => {
						saver.flush();
						previewing = true;
					}}>preview</button
				>
			</div>
		</div>
		{#if previewing}
			<!-- Clicking the text (but not a link in it) switches to editing; the edit button is the keyboard route. -->
			<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
			<div
				class="md cursor-text rounded-[3px] border border-line-soft px-2.5 py-2 leading-relaxed"
				style:min-height="{rows * 1.625 + 1.25}em"
				onclick={(e) => {
					if (!(e.target as Element).closest('a')) editDescription();
				}}
			>
				{#if description.trim()}
					{@html renderMarkdown(description)}
				{:else}
					<p class="text-dim">no description. click to write one.</p>
				{/if}
			</div>
		{:else}
			<textarea
				id="task-desc-{id}"
				bind:value={description}
				{rows}
				class="field h-auto resize-y py-2 leading-relaxed text-text"
				oninput={() => saver.queue({ description })}
				onblur={saver.flush}
			></textarea>
		{/if}
	</div>
{/snippet}

{#snippet priorityField()}
	<div class="flex flex-col gap-1.5" role="group" aria-labelledby="task-pri-{id}">
		<span id="task-pri-{id}" class="label">priority</span>
		<div class="grid grid-cols-4 overflow-hidden rounded-[3px] border border-line-strong">
			{#each priorities as p, i (p.label)}
				<button
					type="button"
					aria-pressed={task!.priority === p.value}
					class="h-7 text-sm max-sm:h-11 max-sm:text-[13px] {i < 3 ? 'border-r border-line-strong' : ''} {task!.priority ===
					p.value
						? p.on
						: 'text-soft hover:text-bright'}"
					onclick={() => saver.save({ priority: p.value })}>{p.label}</button
				>
			{/each}
		</div>
	</div>
{/snippet}

{#snippet projectField()}
	<div class="flex min-w-0 flex-col gap-1.5">
		<label for="task-proj-{id}" class="label">project</label>
		<select
			id="task-proj-{id}"
			class="field min-w-0"
			value={task!.project_id ?? ''}
			onchange={(e) => {
				const v = e.currentTarget.value;
				saver.save({ project_id: v ? Number(v) : null });
			}}
		>
			<option value="">none</option>
			{#each projectOptions as p (p.id)}
				<option value={p.id}>{p.name}</option>
			{/each}
		</select>
	</div>
{/snippet}

{#snippet categoryField()}
	<div class="flex min-w-0 flex-col gap-1.5">
		<label for="task-cat-{id}" class="label"
			>category{#if task!.project_id}<span class="text-faint sm:hidden">{' · from project'}</span>{/if}</label
		>
		{#if task!.project_id}
			{@const c = category(task!.effective_category_id)}
			<div
				id="task-cat-{id}"
				class="flex h-[30px] items-center justify-between rounded-[3px] border border-dashed border-line px-2.5 max-sm:h-11"
			>
				<span class="flex items-center gap-2" style:color={categoryColor(c?.id)}>
					{#if c}<span class="size-[9px]" style:background={c.color}></span>{/if}
					{c?.name ?? 'none'}
				</span>
				<span class="text-xs text-dim max-sm:hidden">from project</span>
			</div>
		{:else}
			<select
				id="task-cat-{id}"
				class="field min-w-0"
				value={task!.category_id ?? ''}
				onchange={(e) => {
					const v = e.currentTarget.value;
					saver.save({ category_id: v ? Number(v) : null });
				}}
			>
				<option value="">none</option>
				{#each app.dash?.categories ?? [] as c (c.id)}
					<option value={c.id}>{c.name}</option>
				{/each}
			</select>
		{/if}
	</div>
{/snippet}

{#snippet scheduledField()}
	<div class="flex min-w-0 flex-col gap-1.5">
		<label for="task-sched-{id}" class="label"
			>scheduled <span class="text-faint max-sm:hidden">· when to work on it</span></label
		>
		<DateField
			id="task-sched-{id}"
			value={task!.scheduled_date}
			quick
			onchange={(v) => saver.save({ scheduled_date: v })}
		/>
	</div>
{/snippet}

{#snippet dueField()}
	<div class="flex min-w-0 flex-col gap-1.5">
		<label for="task-due-{id}" class="label">due <span class="text-faint max-sm:hidden">· deadline</span></label>
		<DateField
			id="task-due-{id}"
			value={task!.due_date}
			tone="var(--color-amber)"
			onchange={(v) => saver.save({ due_date: v })}
		/>
	</div>
{/snippet}

{#snippet deleteButtons()}
	{#if confirmDelete}
		<button type="button" class="btn btn-danger" onclick={remove}>confirm delete</button>
		<button type="button" class="btn" onclick={() => (confirmDelete = false)}>keep</button>
	{:else}
		<button type="button" class="btn btn-danger" onclick={() => (confirmDelete = true)}>delete</button>
	{/if}
{/snippet}
