<script lang="ts">
	import { onDestroy, untrack } from 'svelte';
	import { api } from '$lib/api';
	import { autosave } from '$lib/autosave.svelte';
	import {
		clock,
		dateIn,
		daysLeft,
		percent,
		progressBar,
		short,
		statusColor,
		statusLabel
	} from '$lib/format';
	import {
		app,
		category,
		openModal,
		scheduleRefresh,
		timezone,
		today
	} from '$lib/state.svelte';
	import { projectPace } from '$lib/pace';
	import type { ProjectDetail, ProjectPatch, ProjectStatus, Task } from '$lib/types';
	import Checkbox from './Checkbox.svelte';
	import DateField from './DateField.svelte';
	import Modal from './Modal.svelte';
	import ModalHeader from './ModalHeader.svelte';
	import Pri from './Pri.svelte';

	let { id, onclose }: { id: number; onclose: () => void } = $props();

	let project = $state<ProjectDetail | null>(null);
	let loadError = $state('');
	let name = $state('');
	let notes = $state('');
	let newTask = $state('');
	let showCompleted = $state(true);
	let confirmDelete = $state(false);

	const saver = autosave<ProjectPatch>(async (patch) => {
		project = await api.updateProject(id, patch);
		scheduleRefresh();
	});

	async function load() {
		try {
			const p = await api.getProject(id);
			if (!project) {
				name = p.name;
				notes = p.notes;
			}
			project = p;
		} catch (e) {
			loadError = e instanceof Error ? e.message : String(e);
		}
	}

	// Load now, and again whenever the dashboard reloads (a task modal may have changed our tasks).
	$effect(() => {
		void app.syncedAt;
		untrack(load);
	});

	function close() {
		saver.flush();
		onclose();
	}
	onDestroy(saver.flush);

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

	async function addTask(e: SubmitEvent) {
		e.preventDefault();
		const taskName = newTask.trim();
		if (!taskName) return;
		newTask = '';
		await run(() => api.createTask({ name: taskName, project_id: id }));
	}

	function toggleTask(t: Task) {
		run(() => (t.completed_on ? api.reopenTask(t.id) : api.completeTask(t.id)));
	}

	async function remove() {
		await run(() => api.deleteProject(id));
		if (!saver.status.error) onclose();
	}

	const statuses: ProjectStatus[] = ['todo', 'in_progress', 'waiting', 'complete'];

	const cat = $derived(category(project?.category_id));
	const total = $derived(project ? project.open_count + project.done_count : 0);
	const pct = $derived(project ? percent(project.done_count, total) : 0);
	/** The project pinning this one would replace. */
	const pinnedRival = $derived(
		app.dash?.projects.find((p) => p.pinned && p.category_id === cat?.id && p.id !== id)
	);
	const archived = $derived(!!project?.archived_at);
	const pace = $derived(
		project &&
			projectPace(
				project.completed_tasks.map((t) => t.completed_on!),
				project.open_count,
				today(),
				project.due_date
			)
	);
	const paceMax = $derived(Math.max(1, ...(pace?.weeks.map((w) => w.count) ?? [])));

	const pinHint = $derived.by(() => {
		if (!project) return '';
		if (archived) return "an archived project can't be pinned";
		if (project.pinned) return `pinned in ${cat?.name}; pinning another ${cat?.name} project replaces it`;
		if (!cat) return 'give the project a category to pin it';
		if (project.status === 'complete') return "a completed project can't be pinned";
		return pinnedRival ? `replaces ${pinnedRival.name} as the pinned ${cat.name} project` : `pin in ${cat.name}`;
	});
</script>

<Modal labelledby="proj-name-{id}" width={960} height={780} onclose={close}>
	<ModalHeader status={saver.status} onclose={close}>
		<span class="font-bold text-green">project</span>
		{#if cat}<span class="text-dim">{cat.name}/</span>{/if}
		{#if project?.pinned}<span class="text-amber">◆ pinned</span>{/if}
		{#if archived}<span class="text-muted">▣ archived</span>{/if}
	</ModalHeader>

	{#if loadError}
		<p class="p-4 text-red">E: {loadError}</p>
	{:else if project}
		{#if archived}
			<div class="flex items-center justify-between gap-3 border-b border-line-soft bg-raised px-[18px] py-2 text-xs text-muted">
				<span
					>archived {short(dateIn(project.archived_at!, timezone()))} · hidden from the dashboard, open tasks
					included</span
				>
				<button type="button" class="btn-sm" onclick={() => saver.save({ archived: false })}>unarchive</button>
			</div>
		{/if}
		<div
			class="grid items-end gap-3.5 border-b border-line-soft px-[18px] pt-4 pb-3.5 sm:grid-cols-[minmax(0,1fr)_190px_190px]"
		>
			<div class="flex flex-col gap-1">
				<label for="proj-name-{id}" class="label">name</label>
				<input
					id="proj-name-{id}"
					bind:value={name}
					class="field h-[40px] border-line-strong text-xl font-bold {name.trim() ? '' : 'border-red!'}"
					oninput={() => name.trim() && saver.queue({ name })}
					onkeydown={(e) => e.key === 'Enter' && saver.flush()}
					onblur={saver.flush}
				/>
			</div>
			<div class="flex flex-col gap-1">
				<label for="proj-cat-{id}" class="label">category</label>
				<select
					id="proj-cat-{id}"
					class="field h-[40px]"
					value={project.category_id ?? ''}
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
			</div>
			<div class="flex flex-col gap-1">
				<label for="proj-due-{id}" class="label">
					due
					{#if project.due_date}<span class="text-faint">· {daysLeft(project.due_date, today())}</span
						>{/if}
				</label>
				<DateField
					id="proj-due-{id}"
					value={project.due_date}
					height={40}
					clearButton={false}
					onchange={(v) => saver.save({ due_date: v })}
				/>
			</div>
		</div>

		<div
			class="flex flex-wrap items-center justify-between gap-4 border-b border-line-soft px-[18px] py-3"
		>
			<div class="flex flex-col gap-1.5" role="group" aria-labelledby="proj-status-{id}">
				<span id="proj-status-{id}" class="label">status</span>
				<div class="flex overflow-hidden rounded-[3px] border border-line-strong">
					{#each statuses as s, i (s)}
						{@const on = project.status === s}
						<button
							type="button"
							aria-pressed={on}
							class="h-[30px] px-3 text-sm {i < 3 ? 'border-r border-line-strong' : ''} {on
								? 'bg-green-deep font-bold'
								: 'text-soft hover:text-bright'}"
							style:color={on ? statusColor[s] : undefined}
							onclick={() => saver.save({ status: s })}>{statusLabel[s]}</button
						>
					{/each}
				</div>
			</div>

			<div class="flex flex-col gap-1.5">
				<span class="label">priority</span>
				<button
					type="button"
					aria-pressed={project.pinned}
					disabled={!cat || project.status === 'complete' || archived}
					title={pinHint}
					class="flex h-[30px] items-center gap-2 rounded-[3px] border px-2.5 text-sm disabled:cursor-not-allowed disabled:opacity-50 {project.pinned
						? 'border-amber-edge bg-amber-deep text-amber'
						: 'border-line text-soft hover:text-bright'}"
					onclick={() => saver.save({ pinned: !project!.pinned })}
				>
					<span
						class="flex h-3.5 w-[26px] rounded-full p-0.5 {project.pinned
							? 'justify-end bg-amber'
							: 'justify-start bg-line-strong'}"
						><span class="size-2.5 rounded-full bg-modal"></span></span
					>
					{project.pinned ? `◆ pinned in ${cat?.name}` : cat ? `pin in ${cat.name}` : 'pin (needs category)'}
				</button>
			</div>

			<div class="flex min-w-[190px] flex-col gap-1">
				<span class="flex justify-between text-xs text-dim"
					><span>progress</span><span class="text-text">{project.done_count}/{total} · {pct}%</span></span
				>
				<span class="tracking-[-0.5px] text-green">{progressBar(pct)}</span>
			</div>
		</div>

		<div class="grid min-h-0 grow grid-cols-1 overflow-y-auto sm:grid-cols-[minmax(0,1fr)_300px]">
			<div class="flex min-h-0 flex-col gap-0.5 border-line-soft px-[18px] py-3 sm:overflow-y-auto sm:border-r">
				<form
					onsubmit={addTask}
					class="mb-2 flex h-8 shrink-0 items-center gap-2 rounded-[3px] border border-line-strong bg-raised px-2.5"
				>
					<span class="font-bold text-green" aria-hidden="true">❯</span>
					<input
						bind:value={newTask}
						aria-label="Add task to {project.name}"
						placeholder="add task to {project.name}, enter to save"
						class="min-w-0 grow bg-transparent text-bright outline-none"
					/>
				</form>

				<div
					class="grid grid-cols-[18px_24px_minmax(0,1fr)_52px_52px_40px] gap-2 border-b border-line-soft pb-1 text-xs text-dim"
				>
					<span></span><span>PRI</span><span>OPEN · {project.open_count}</span><span>SCHED</span><span
						>DUE</span
					><span>SUB</span>
				</div>
				{#each project.open_tasks as t (t.id)}
					<div
						class="grid grid-cols-[18px_24px_minmax(0,1fr)_52px_52px_40px] items-center gap-2 border-b border-line-soft py-[7px]"
					>
						<Checkbox checked={false} size={14} label="Complete {t.name}" onclick={() => toggleTask(t)} />
						<button
							type="button"
							class="col-span-5 grid grid-cols-subgrid items-center text-left hover:bg-hover"
							onclick={() => openModal({ kind: 'task', id: t.id })}
						>
							<Pri priority={t.priority} />
							<span class="truncate text-bright">{t.name}</span>
							<span class="text-muted">{t.scheduled_date === today() ? 'today' : short(t.scheduled_date)}</span>
							<span class={t.due_date && t.due_date <= today() ? 'text-amber' : 'text-muted'}
								>{t.due_date === today() ? 'today' : short(t.due_date)}</span
							>
							<span class="text-muted"
								>{t.subtasks_total ? `${t.subtasks_done}/${t.subtasks_total}` : '—'}</span
							>
						</button>
					</div>
				{:else}
					<p class="py-2 text-dim">no open tasks</p>
				{/each}

				<button
					type="button"
					class="flex justify-between border-b border-line-soft pt-3 pb-1 text-xs"
					aria-expanded={showCompleted}
					onclick={() => (showCompleted = !showCompleted)}
				>
					<span class="text-dim">{showCompleted ? '▾' : '▸'} COMPLETED · {project.done_count}</span>
					<span class="text-faint">newest first</span>
				</button>
				{#if showCompleted}
					{#each project.completed_tasks as t (t.id)}
						<div
							class="grid grid-cols-[18px_minmax(0,1fr)_100px] items-center gap-2 border-b border-line-soft py-1.5 text-dim"
						>
							<Checkbox checked size={14} label="Reopen {t.name}" onclick={() => toggleTask(t)} />
							<button
								type="button"
								class="col-span-2 grid grid-cols-subgrid items-center text-left hover:bg-hover"
								onclick={() => openModal({ kind: 'task', id: t.id })}
							>
								<span class="truncate line-through">{t.name}</span>
								<span class="text-right text-xs"
									>{short(t.completed_on)} {t.completed_at ? clock(t.completed_at, timezone()) : ''}</span
								>
							</button>
						</div>
					{/each}
				{/if}
			</div>

			<div class="flex flex-col gap-1.5 bg-side px-4 py-3">
				<label for="proj-notes-{id}" class="label">notes <span class="text-faint">· optional</span></label>
				<textarea
					id="proj-notes-{id}"
					bind:value={notes}
					class="field h-auto min-h-40 grow resize-none py-2 text-sm leading-relaxed text-text"
					oninput={() => saver.queue({ notes })}
					onblur={saver.flush}
				></textarea>

				{#if pace}
					<div class="mt-2 flex shrink-0 flex-col gap-1.5" role="group" aria-labelledby="proj-pace-{id}">
						<span id="proj-pace-{id}" class="label">completions · last 6 weeks</span>
						<div class="flex h-11 items-end gap-1 border-b border-line">
							{#each pace.weeks as w, i (w.to)}
								<span
									class="grow {i === pace.weeks.length - 1 ? 'bg-green' : 'bg-heat-2'}"
									style:height="{(w.count / paceMax) * 100}%"
									title="{short(w.from)}–{short(w.to)} · {w.count} completed"
								><span class="sr-only">{short(w.from)} to {short(w.to)}: {w.count} completed</span></span>
							{/each}
						</div>
						<span
							class="text-xs {pace.onTrack === false ? 'text-amber' : 'text-muted'}"
							title={pace.estimate ? `estimated done ${pace.estimate} at this pace` : undefined}
							>{pace.summary}</span
						>
					</div>
				{/if}
			</div>
		</div>

		<div
			class="flex h-[52px] shrink-0 items-center justify-between gap-2 border-t border-line bg-raised px-3.5"
		>
			<div class="flex items-center gap-2">
				{#if confirmDelete}
					<button type="button" class="btn btn-danger" onclick={remove}>confirm delete</button>
					<span class="hidden text-xs text-dim sm:inline">tasks are kept as standalone</span>
					<button type="button" class="btn" onclick={() => (confirmDelete = false)}>keep</button>
				{:else}
					<button type="button" class="btn btn-danger" onclick={() => (confirmDelete = true)}
						>delete project</button
					>
					<button
						type="button"
						class="btn"
						title={archived ? 'show it and its open tasks on the dashboard again' : 'hide it and its open tasks from the dashboard'}
						onclick={() => saver.save({ archived: !archived })}>{archived ? 'unarchive' : 'archive'}</button
					>
				{/if}
			</div>
			<div class="flex items-center gap-2.5">
				<span class="hidden text-xs text-dim sm:inline"><span class="key">esc</span> close</span>
				{#if project.status === 'complete'}
					<button type="button" class="btn" onclick={() => saver.save({ status: 'in_progress' })}
						>reopen project</button
					>
				{:else}
					<button
						type="button"
						class="btn btn-primary"
						onclick={() => saver.save({ status: 'complete' })}>mark project complete</button
					>
				{/if}
			</div>
		</div>
	{:else}
		<p class="p-4 text-dim">loading…</p>
	{/if}
</Modal>
