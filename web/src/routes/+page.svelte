<script lang="ts">
	import { onMount } from 'svelte';
	import AddTask from '$lib/components/AddTask.svelte';
	import CategoryModal from '$lib/components/CategoryModal.svelte';
	import Header from '$lib/components/Header.svelte';
	import HistoryModal from '$lib/components/HistoryModal.svelte';
	import MobileToday from '$lib/components/MobileToday.svelte';
	import PinnedPanel from '$lib/components/PinnedPanel.svelte';
	import ProjectModal from '$lib/components/ProjectModal.svelte';
	import ProjectsPanel from '$lib/components/ProjectsPanel.svelte';
	import QueuePanel from '$lib/components/QueuePanel.svelte';
	import StatsPanel from '$lib/components/StatsPanel.svelte';
	import StatsStrip from '$lib/components/StatsStrip.svelte';
	import StatusBar from '$lib/components/StatusBar.svelte';
	import TaskModal from '$lib/components/TaskModal.svelte';
	import TodayPanel from '$lib/components/TodayPanel.svelte';
	import { desktop } from '$lib/media.svelte';
	import { app, closeModal, refresh } from '$lib/state.svelte';

	// Keep the dashboard current: poll every minute and reload when the tab regains focus.
	onMount(() => {
		refresh();
		const timer = setInterval(refresh, 60_000);
		window.addEventListener('focus', refresh);
		return () => {
			clearInterval(timer);
			window.removeEventListener('focus', refresh);
		};
	});
</script>

<div class="flex min-h-dvh flex-col bg-bg lg:h-dvh lg:overflow-hidden">
	<Header />

	{#if app.dash && desktop.current}
		<main
			class="grid min-h-0 grow grid-cols-[380px_minmax(0,1fr)_320px] gap-x-3 gap-y-[18px] px-3.5 pt-[18px] pb-3"
		>
			<TodayPanel />
			<div class="flex min-h-0 flex-col gap-[18px]">
				<PinnedPanel />
				<QueuePanel />
			</div>
			<div class="flex min-h-0 flex-col gap-[18px]">
				<StatsPanel />
				<ProjectsPanel />
			</div>
		</main>
	{:else if app.dash}
		<!-- Mobile: one scrolling column, projects last so every project stays reachable. -->
		<main class="flex grow flex-col gap-3.5 p-3 pb-5">
			<StatsStrip />
			<AddTask large />
			<MobileToday />
			<PinnedPanel scroll />
			<QueuePanel compact />
			<ProjectsPanel />
		</main>
	{:else}
		<main class="grow p-6 text-dim">
			{app.error ? '' : 'connecting…'}
			{#if app.error}<span class="text-red">E: can't reach the API ({app.error})</span>{/if}
		</main>
	{/if}

	<StatusBar />
</div>

{#each app.modals as m, i (i + m.kind + ('id' in m ? m.id : ''))}
	{#if m.kind === 'task'}
		<TaskModal id={m.id} onclose={() => closeModal(i)} />
	{:else if m.kind === 'project'}
		<ProjectModal id={m.id} onclose={() => closeModal(i)} />
	{:else if m.kind === 'history'}
		<HistoryModal day={m.day} onclose={() => closeModal(i)} />
	{:else}
		<CategoryModal onclose={() => closeModal(i)} />
	{/if}
{/each}
