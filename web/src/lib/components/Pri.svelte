<script lang="ts">
	import { priorityColor } from '$lib/format';
	import type { Priority } from '$lib/types';

	let { priority }: { priority: Priority | null } = $props();
</script>

{#if priority}
	<span class="font-bold" style:color={priorityColor[priority]}>P{priority}</span>
{:else}
	<span class="text-faint" aria-label="no priority">·</span>
{/if}
